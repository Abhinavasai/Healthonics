import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { SimpleChange } from '@angular/core';
import { ContextualCommentsPanelComponent } from './contextual-comments-panel.component';
import { AuthService } from '../../services/auth.service';

describe('ContextualCommentsPanelComponent', () => {
  let fixture: ComponentFixture<ContextualCommentsPanelComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ContextualCommentsPanelComponent],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([]), AuthService]
    }).compileComponents();

    localStorage.setItem('user', JSON.stringify({ id: 'u1', email: 'admin@demo.com', role: 'admin' }));
    fixture = TestBed.createComponent(ContextualCommentsPanelComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.componentInstance.contextType = 'document';
    fixture.componentInstance.contextId = 'd1';
    fixture.componentInstance.ngOnChanges({
      contextType: new SimpleChange(undefined, 'document', true),
      contextId: new SimpleChange(undefined, 'd1', true)
    });
    fixture.detectChanges();
  });

  afterEach(() => {
    localStorage.removeItem('user');
    httpMock.verify();
  });

  it('loads comments for context', () => {
    httpMock.expectOne('/api/comments/document/d1').flush({
      comments: []
    });
    fixture.detectChanges();
    expect((fixture.nativeElement as HTMLElement).querySelector('[data-cy="context-comments-panel"]')).toBeTruthy();
  });

  it('creates a comment and reloads list', () => {
    httpMock.expectOne('/api/comments/document/d1').flush({
      comments: []
    });
    fixture.componentInstance.newBody = 'Follow-up required';
    fixture.componentInstance.newVisibility = 'care_team';
    fixture.componentInstance.create();

    const createReq = httpMock.expectOne('/api/comments/document/d1');
    expect(createReq.request.method).toBe('POST');
    createReq.flush({ id: 'c1', visibility: 'care_team' });

    httpMock.expectOne('/api/comments/document/d1').flush({
      comments: [
        {
          id: 'c1',
          context_type: 'document',
          context_id: 'd1',
          author_user_id: 'u1',
          body: 'Follow-up required',
          visibility: 'care_team',
          created_at: 'now'
        }
      ]
    });
    expect(fixture.componentInstance.comments.length).toBe(1);
  });
});

