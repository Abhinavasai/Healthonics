package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/healthonyx/backend/db"
)

type embeddingChunkInput struct {
	ChunkIndex int       `json:"chunk_index"`
	Body       string    `json:"body"`
	Embedding  []float64 `json:"embedding"`
}

type embeddingsUpsertBody struct {
	Chunks []embeddingChunkInput `json:"chunks" binding:"required"`
}

type loadedChunk struct {
	DocID      uuid.UUID
	ChunkIndex int
	Body       string
	Title      string
	Vec        []float64
}

type similarityHit struct {
	Similarity   float64 `json:"similarity"`
	DocumentAID  string  `json:"document_a_id"`
	DocumentBID  string  `json:"document_b_id"`
	TitleA       string  `json:"title_a"`
	TitleB       string  `json:"title_b"`
	ChunkAIndex  int     `json:"chunk_a_index"`
	ChunkBIndex  int     `json:"chunk_b_index"`
	ExcerptA     string  `json:"excerpt_a"`
	ExcerptB     string  `json:"excerpt_b"`
	ConflictHint string  `json:"conflict_hint"`
}

func excerptKB(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// UpsertEmbeddings replaces all chunks + embeddings for a knowledge doc (admin embedding pipeline).
func (h *KnowledgeAdminHandler) UpsertEmbeddings(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	docID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	var body embeddingsUpsertBody
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Chunks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunks required"})
		return
	}
	if len(body.Chunks) > 256 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max 256 chunks"})
		return
	}

	dim := -1
	for _, ch := range body.Chunks {
		if ch.ChunkIndex < 0 || strings.TrimSpace(ch.Body) == "" || len(ch.Embedding) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "each chunk needs chunk_index, body, embedding (min length 4)"})
			return
		}
		if dim < 0 {
			dim = len(ch.Embedding)
		} else if len(ch.Embedding) != dim {
			c.JSON(http.StatusBadRequest, gin.H{"error": "all embeddings must share the same dimension"})
			return
		}
	}

	ctx := c.Request.Context()
	var exists bool
	if err := db.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM knowledge_docs WHERE id = $1)`, docID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM knowledge_doc_chunks WHERE document_id = $1`, docID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	for _, ch := range body.Chunks {
		payload, err := json.Marshal(ch.Embedding)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid embedding"})
			return
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO knowledge_doc_chunks (document_id, chunk_index, body, embedding)
			VALUES ($1, $2, $3, $4::jsonb)
		`, docID, ch.ChunkIndex, strings.TrimSpace(ch.Body), payload); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"document_id": docID.String(), "chunks": len(body.Chunks), "embedding_dim": dim})
}

// SimilarityScan compares embeddings across different KB documents (pairwise). Used for overlap / conflict review.
func (h *KnowledgeAdminHandler) SimilarityScan(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	threshold := 0.85
	if v := strings.TrimSpace(c.Query("threshold")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 && f <= 1 {
			threshold = f
		}
	}
	maxRows := 1200
	if v := strings.TrimSpace(c.Query("max_chunks")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 10 && n <= 5000 {
			maxRows = n
		}
	}

	ctx := c.Request.Context()
	rows, err := db.Pool.Query(ctx, `
		SELECT c.document_id, c.chunk_index, c.body, c.embedding, d.title
		FROM knowledge_doc_chunks c
		INNER JOIN knowledge_docs d ON d.id = c.document_id
		ORDER BY d.updated_at DESC, c.document_id, c.chunk_index
		LIMIT $1
	`, maxRows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	var chunks []loadedChunk
	for rows.Next() {
		var lc loadedChunk
		var embRaw []byte
		if err := rows.Scan(&lc.DocID, &lc.ChunkIndex, &lc.Body, &embRaw, &lc.Title); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		if err := json.Unmarshal(embRaw, &lc.Vec); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "stored embedding corrupt"})
			return
		}
		chunks = append(chunks, lc)
	}

	var hits []similarityHit
	for i := 0; i < len(chunks); i++ {
		for j := i + 1; j < len(chunks); j++ {
			if chunks[i].DocID == chunks[j].DocID {
				continue
			}
			if len(chunks[i].Vec) != len(chunks[j].Vec) || len(chunks[i].Vec) == 0 {
				continue
			}
			sim := CosineSimilarity(chunks[i].Vec, chunks[j].Vec)
			if sim < threshold {
				continue
			}

			a, b := chunks[i], chunks[j]
			if strings.Compare(a.DocID.String(), b.DocID.String()) > 0 {
				a, b = b, a
			}

			ta := strings.TrimSpace(strings.ToLower(a.Title))
			tb := strings.TrimSpace(strings.ToLower(b.Title))
			hint := "high_similarity"
			if ta != tb && sim >= 0.92 {
				hint = "potential_conflict_review"
			}
			if ta == tb && sim >= 0.95 {
				hint = "near_duplicate_under_same_title"
			}

			hits = append(hits, similarityHit{
				Similarity:   sim,
				DocumentAID:  a.DocID.String(),
				DocumentBID:  b.DocID.String(),
				TitleA:       a.Title,
				TitleB:       b.Title,
				ChunkAIndex:  a.ChunkIndex,
				ChunkBIndex:  b.ChunkIndex,
				ExcerptA:     excerptKB(a.Body, 180),
				ExcerptB:     excerptKB(b.Body, 180),
				ConflictHint: hint,
			})
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].Similarity > hits[j].Similarity
	})
	if len(hits) > 200 {
		hits = hits[:200]
	}

	c.JSON(http.StatusOK, gin.H{
		"threshold":      threshold,
		"chunks_scanned": len(chunks),
		"pairs":          hits,
	})
}
