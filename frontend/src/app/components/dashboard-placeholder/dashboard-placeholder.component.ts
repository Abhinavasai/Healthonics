import { Component, Input, OnInit } from '@angular/core';
import { AsyncPipe, TitleCasePipe, DatePipe, LowerCasePipe } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { BootstrapService } from '../../services/bootstrap.service';

@Component({
  selector: 'app-dashboard-placeholder',
  standalone: true,
  imports: [AsyncPipe, TitleCasePipe, DatePipe, LowerCasePipe],
  template: `
    <div class="dashboard">
      <!-- Header -->
      <div class="dashboard-header">
        <div class="header-content">
          <h1 class="dashboard-title">
            <span class="title-icon">
              @if (roleTitle === 'Patient') { 🧑‍⚕️ }
              @if (roleTitle === 'Doctor') { 👨‍⚕️ }
              @if (roleTitle === 'Admin') { 🛡️ }
            </span>
            {{ roleTitle }} Dashboard
          </h1>
          <p class="dashboard-date">{{ today | date:'EEEE, MMMM d, y' }}</p>
          <p class="dashboard-date">{{ appName }} · API {{ apiVersion }}</p>
        </div>
      </div>

      <!-- Stats Cards -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon stat-icon-cyan">📊</div>
          <div class="stat-info">
            <span class="stat-value">24</span>
            <span class="stat-label">Total Records</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon stat-icon-purple">📅</div>
          <div class="stat-info">
            <span class="stat-value">3</span>
            <span class="stat-label">Appointments</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon stat-icon-green">✓</div>
          <div class="stat-info">
            <span class="stat-value">12</span>
            <span class="stat-label">Completed</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon stat-icon-pink">⏳</div>
          <div class="stat-info">
            <span class="stat-value">5</span>
            <span class="stat-label">Pending</span>
          </div>
        </div>
      </div>

      <!-- Main Content -->
      <div class="content-grid">
        <div class="content-card main-card">
          <div class="card-header">
            <h2>Recent Activity</h2>
            <span class="card-badge">Live</span>
          </div>
          <div class="card-content">
            <div class="activity-item">
              <div class="activity-dot"></div>
              <div class="activity-info">
                <span class="activity-text">System initialized successfully</span>
                <span class="activity-time">Just now</span>
              </div>
            </div>
            <div class="activity-item">
              <div class="activity-dot"></div>
              <div class="activity-info">
                <span class="activity-text">Dashboard loaded</span>
                <span class="activity-time">2 minutes ago</span>
              </div>
            </div>
            <div class="activity-item">
              <div class="activity-dot"></div>
              <div class="activity-info">
                <span class="activity-text">User authenticated</span>
                <span class="activity-time">5 minutes ago</span>
              </div>
            </div>
          </div>
        </div>
        
        <div class="content-card">
          <div class="card-header">
            <h2>Quick Actions</h2>
          </div>
          <div class="card-content">
            <button class="action-btn">
              <span class="action-icon">➕</span>
              <span>New Record</span>
            </button>
            <button class="action-btn">
              <span class="action-icon">📋</span>
              <span>View Reports</span>
            </button>
            <button class="action-btn">
              <span class="action-icon">⚙️</span>
              <span>Settings</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Coming Soon Notice -->
      <div class="coming-soon">
        <div class="coming-soon-content">
          <span class="coming-soon-icon">🚀</span>
          <h3>More Features Coming Soon</h3>
          <p>We're working on exciting new features for your {{ roleTitle | lowercase }} dashboard.</p>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .dashboard {
      max-width: 1400px;
      margin: 0 auto;
    }
    
    .dashboard-header {
      margin-bottom: var(--space-2xl);
    }
    
    .dashboard-title {
      display: flex;
      align-items: center;
      gap: var(--space-md);
      font-size: 2rem;
      font-weight: 700;
      color: var(--text-primary);
      margin: 0 0 var(--space-sm);
      
      .title-icon {
        font-size: 2.5rem;
      }
    }
    
    .dashboard-date {
      color: var(--text-muted);
      font-size: 0.95rem;
    }
    
    /* Stats Grid */
    .stats-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: var(--space-lg);
      margin-bottom: var(--space-2xl);
    }
    
    .stat-card {
      display: flex;
      align-items: center;
      gap: var(--space-lg);
      padding: var(--space-xl);
      background: var(--gradient-card);
      border: var(--border-subtle);
      border-radius: var(--radius-lg);
      transition: var(--transition-base);
      
      &:hover {
        transform: translateY(-4px);
        border-color: rgba(255, 255, 255, 0.2);
      }
    }
    
    .stat-icon {
      width: 50px;
      height: 50px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: var(--radius-md);
      font-size: 1.5rem;
      
      &.stat-icon-cyan { background: rgba(0, 240, 255, 0.2); }
      &.stat-icon-purple { background: rgba(123, 97, 255, 0.2); }
      &.stat-icon-green { background: rgba(0, 255, 136, 0.2); }
      &.stat-icon-pink { background: rgba(255, 0, 170, 0.2); }
    }
    
    .stat-info {
      display: flex;
      flex-direction: column;
    }
    
    .stat-value {
      font-family: 'Space Grotesk', sans-serif;
      font-size: 1.75rem;
      font-weight: 700;
      color: var(--text-primary);
    }
    
    .stat-label {
      font-size: 0.85rem;
      color: var(--text-muted);
    }
    
    /* Content Grid */
    .content-grid {
      display: grid;
      grid-template-columns: 2fr 1fr;
      gap: var(--space-xl);
      margin-bottom: var(--space-2xl);
      
      @media (max-width: 900px) {
        grid-template-columns: 1fr;
      }
    }
    
    .content-card {
      background: var(--gradient-card);
      border: var(--border-subtle);
      border-radius: var(--radius-lg);
      overflow: hidden;
    }
    
    .card-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: var(--space-lg) var(--space-xl);
      border-bottom: var(--border-subtle);
      
      h2 {
        font-size: 1.1rem;
        font-weight: 600;
        margin: 0;
      }
    }
    
    .card-badge {
      display: flex;
      align-items: center;
      gap: var(--space-xs);
      padding: var(--space-xs) var(--space-sm);
      background: rgba(0, 255, 136, 0.1);
      border: 1px solid rgba(0, 255, 136, 0.3);
      border-radius: var(--radius-full);
      font-size: 0.7rem;
      color: var(--color-success);
      text-transform: uppercase;
      letter-spacing: 0.05em;
      
      &::before {
        content: '';
        width: 6px;
        height: 6px;
        background: var(--color-success);
        border-radius: 50%;
        animation: pulse 2s ease-in-out infinite;
      }
    }
    
    .card-content {
      padding: var(--space-xl);
    }
    
    /* Activity Items */
    .activity-item {
      display: flex;
      align-items: flex-start;
      gap: var(--space-md);
      padding: var(--space-md) 0;
      border-bottom: 1px solid rgba(255, 255, 255, 0.05);
      
      &:last-child {
        border-bottom: none;
      }
    }
    
    .activity-dot {
      width: 8px;
      height: 8px;
      background: var(--color-primary);
      border-radius: 50%;
      margin-top: 6px;
    }
    
    .activity-info {
      display: flex;
      flex-direction: column;
      gap: var(--space-xs);
    }
    
    .activity-text {
      color: var(--text-primary);
      font-size: 0.95rem;
    }
    
    .activity-time {
      color: var(--text-muted);
      font-size: 0.8rem;
    }
    
    /* Action Buttons */
    .action-btn {
      width: 100%;
      display: flex;
      align-items: center;
      gap: var(--space-md);
      padding: var(--space-md) var(--space-lg);
      margin-bottom: var(--space-sm);
      background: rgba(255, 255, 255, 0.03);
      border: 1px solid rgba(255, 255, 255, 0.1);
      border-radius: var(--radius-md);
      color: var(--text-secondary);
      font-size: 0.9rem;
      cursor: pointer;
      transition: var(--transition-fast);
      
      &:hover {
        background: rgba(0, 240, 255, 0.1);
        border-color: var(--color-primary);
        color: var(--color-primary);
      }
      
      &:last-child {
        margin-bottom: 0;
      }
    }
    
    .action-icon {
      font-size: 1.25rem;
    }
    
    /* Coming Soon */
    .coming-soon {
      background: var(--gradient-card);
      border: var(--border-subtle);
      border-radius: var(--radius-lg);
      padding: var(--space-3xl);
      text-align: center;
    }
    
    .coming-soon-content {
      max-width: 400px;
      margin: 0 auto;
    }
    
    .coming-soon-icon {
      font-size: 3rem;
      display: block;
      margin-bottom: var(--space-lg);
    }
    
    .coming-soon h3 {
      font-size: 1.25rem;
      margin: 0 0 var(--space-sm);
    }
    
    .coming-soon p {
      color: var(--text-muted);
      margin: 0;
    }
    
    @keyframes pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.5; }
    }
  `]
})
export class DashboardPlaceholderComponent implements OnInit {
  @Input() title = 'Patient';
  today = new Date();
  appName = 'Healthonyx';
  apiVersion = 'v1';
  
  constructor(
    public route: ActivatedRoute,
    private bootstrap: BootstrapService
  ) {}

  ngOnInit(): void {
    this.bootstrap.get().subscribe({
      next: (cfg) => {
        this.appName = cfg.app_name || this.appName;
        this.apiVersion = cfg.api_version || this.apiVersion;
      }
    });
  }
  
  get roleTitle(): string {
    return this.title;
  }
}
