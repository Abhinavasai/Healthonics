import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { AdminAuditComponent } from './admin-audit.component';

describe('AdminAuditComponent', () => {
  let fixture: ComponentFixture<AdminAuditComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AdminAuditComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    }).compileComponents();
    fixture = TestBed.createComponent(AdminAuditComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/admin/audit-log').flush({ entries: [] });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('shows empty', () => {
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="admin-audit-empty"]')
    ).toBeTruthy();
  });
});
