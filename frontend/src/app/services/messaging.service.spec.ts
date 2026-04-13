import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { MessagingService } from './messaging.service';

describe('MessagingService', () => {
  let service: MessagingService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [MessagingService]
    });
    service = TestBed.inject(MessagingService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('lists threads and updates unread from thread rows', (done) => {
    service.listThreads().subscribe((threads) => {
      expect(threads.length).toBe(1);
      expect(threads[0].peer_email).toBe('doc@clinic.com');
      expect(service.getUnreadSnapshot()).toBe(2);
      done();
    });
    const req = httpMock.expectOne('/api/messages/threads');
    expect(req.request.method).toBe('GET');
    req.flush({
      threads: [
        {
          id: 't1',
          peer_user_id: 'd1',
          peer_email: 'doc@clinic.com',
          last_message_at: '2026-01-01T12:00:00Z',
          last_preview: 'Hi',
          unread_count: 2
        }
      ]
    });
  });

  it('refreshUnread uses /api/messages/unread', (done) => {
    service.refreshUnread().subscribe((n) => {
      expect(n).toBe(5);
      expect(service.getUnreadSnapshot()).toBe(5);
      done();
    });
    const req = httpMock.expectOne('/api/messages/unread');
    req.flush({ unread_total: 5 });
  });

  it('sendMessage posts to thread messages', () => {
    service.sendMessage('tid', 'hello').subscribe();
    const req = httpMock.expectOne('/api/messages/threads/tid/messages');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ body: 'hello' });
    req.flush({
      id: 'm1',
      thread_id: 'tid',
      sender_id: 'u1',
      body: 'hello',
      created_at: '2026-01-01T12:00:00Z'
    });
  });

  it('createThread posts peer and body', () => {
    service.createThread('peer-uuid', 'first').subscribe();
    const req = httpMock.expectOne('/api/messages/threads');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ peer_user_id: 'peer-uuid', body: 'first' });
    req.flush({
      id: 'new',
      peer_user_id: 'peer-uuid',
      peer_email: 'x@y.com',
      last_message_at: '2026-01-01T12:00:00Z',
      last_preview: 'first',
      unread_count: 0
    });
  });
});
