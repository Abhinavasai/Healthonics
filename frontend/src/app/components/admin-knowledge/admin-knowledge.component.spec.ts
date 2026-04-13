import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { AdminKnowledgeComponent } from './admin-knowledge.component';

describe('AdminKnowledgeComponent', () => {
  let fixture: ComponentFixture<AdminKnowledgeComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AdminKnowledgeComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    }).compileComponents();
    fixture = TestBed.createComponent(AdminKnowledgeComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/admin/knowledge-docs').flush({ documents: [] });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('shows empty', () => {
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="knowledge-empty"]')
    ).toBeTruthy();
  });
});
