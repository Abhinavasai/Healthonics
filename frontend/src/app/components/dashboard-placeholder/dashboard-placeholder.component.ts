import { Component, Input } from '@angular/core';
import { AsyncPipe } from '@angular/common';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-dashboard-placeholder',
  standalone: true,
  imports: [AsyncPipe],
  template: `
    <div class="placeholder">
      <h1>{{ ((route.data | async)?.['title']) ?? title }} Dashboard</h1>
      <p>Coming soon</p>
    </div>
  `,
  styles: [`
    .placeholder {
      padding: 2rem;
      text-align: center;
    }
    h1 { font-size: 1.5rem; }
    p { color: #666; }
  `]
})
export class DashboardPlaceholderComponent {
  @Input() title = 'Patient';
  constructor(public route: ActivatedRoute) {}
}
