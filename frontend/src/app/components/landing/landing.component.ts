import { Component, OnInit, OnDestroy, ElementRef, ViewChild, AfterViewInit, PLATFORM_ID, Inject } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { RouterLink } from '@angular/router';
import { gsap } from 'gsap';
import * as THREE from 'three';

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './landing.component.html',
  styleUrl: './landing.component.scss'
})
export class LandingComponent implements OnInit, AfterViewInit, OnDestroy {
  @ViewChild('canvasContainer') canvasContainer!: ElementRef<HTMLDivElement>;
  
  private scene!: THREE.Scene;
  private camera!: THREE.PerspectiveCamera;
  private renderer!: THREE.WebGLRenderer;
  private particles!: THREE.Points;
  private animationId!: number;
  private isBrowser: boolean;
  private mouseX = 0;
  private mouseY = 0;

  features = [
    {
      icon: '🏥',
      title: 'Smart Health Monitoring',
      description: 'AI-powered real-time health tracking with predictive analytics and early warning systems.'
    },
    {
      icon: '📋',
      title: 'Digital Medical Records',
      description: 'Secure, encrypted storage for all your medical history accessible anytime, anywhere.'
    },
    {
      icon: '👨‍⚕️',
      title: 'Doctor Portal',
      description: 'Comprehensive dashboard for healthcare providers to manage patients efficiently.'
    },
    {
      icon: '📅',
      title: 'Smart Appointments',
      description: 'AI-assisted scheduling with automated reminders and virtual consultation support.'
    },
    {
      icon: '💊',
      title: 'Prescription Management',
      description: 'Digital prescriptions with drug interaction checks and refill automation.'
    },
    {
      icon: '🔒',
      title: 'Enterprise Security',
      description: 'HIPAA-compliant infrastructure with end-to-end encryption and audit trails.'
    }
  ];

  stats = [
    { value: '10K+', label: 'Active Users' },
    { value: '500+', label: 'Healthcare Providers' },
    { value: '99.9%', label: 'Uptime' },
    { value: '24/7', label: 'Support' }
  ];

  constructor(@Inject(PLATFORM_ID) platformId: Object) {
    this.isBrowser = isPlatformBrowser(platformId);
  }

  ngOnInit(): void {
    if (this.isBrowser) {
      window.addEventListener('mousemove', this.onMouseMove.bind(this));
    }
  }

  ngAfterViewInit(): void {
    if (this.isBrowser) {
      this.initThreeJS();
      this.initAnimations();
    }
  }

  ngOnDestroy(): void {
    if (this.isBrowser) {
      window.removeEventListener('mousemove', this.onMouseMove.bind(this));
      if (this.animationId) {
        cancelAnimationFrame(this.animationId);
      }
      if (this.renderer) {
        this.renderer.dispose();
      }
    }
  }

  private onMouseMove(event: MouseEvent): void {
    this.mouseX = (event.clientX / window.innerWidth) * 2 - 1;
    this.mouseY = -(event.clientY / window.innerHeight) * 2 + 1;
  }

  private initThreeJS(): void {
    const container = this.canvasContainer.nativeElement;
    
    // Scene
    this.scene = new THREE.Scene();
    
    // Camera
    this.camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
    this.camera.position.z = 50;
    
    // Renderer
    this.renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
    this.renderer.setSize(window.innerWidth, window.innerHeight);
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    container.appendChild(this.renderer.domElement);
    
    // Create particles
    this.createParticles();
    this.createDNAHelix();
    
    // Handle resize
    window.addEventListener('resize', this.onResize.bind(this));
    
    // Start animation loop
    this.animate();
  }

  private createParticles(): void {
    const particlesGeometry = new THREE.BufferGeometry();
    const particlesCount = 2000;
    const posArray = new Float32Array(particlesCount * 3);
    const colorsArray = new Float32Array(particlesCount * 3);
    
    for (let i = 0; i < particlesCount * 3; i += 3) {
      posArray[i] = (Math.random() - 0.5) * 150;
      posArray[i + 1] = (Math.random() - 0.5) * 150;
      posArray[i + 2] = (Math.random() - 0.5) * 150;
      
      // Cyan to purple gradient
      const t = Math.random();
      colorsArray[i] = t * 0.48 + (1 - t) * 0;       // R
      colorsArray[i + 1] = t * 0.38 + (1 - t) * 0.94; // G
      colorsArray[i + 2] = t * 1 + (1 - t) * 1;       // B
    }
    
    particlesGeometry.setAttribute('position', new THREE.BufferAttribute(posArray, 3));
    particlesGeometry.setAttribute('color', new THREE.BufferAttribute(colorsArray, 3));
    
    const particlesMaterial = new THREE.PointsMaterial({
      size: 0.3,
      vertexColors: true,
      transparent: true,
      opacity: 0.8,
      blending: THREE.AdditiveBlending
    });
    
    this.particles = new THREE.Points(particlesGeometry, particlesMaterial);
    this.scene.add(this.particles);
  }

  private createDNAHelix(): void {
    const helixGroup = new THREE.Group();
    const points = 100;
    const radius = 8;
    const height = 60;
    
    for (let i = 0; i < points; i++) {
      const t = i / points;
      const angle = t * Math.PI * 6;
      
      // First strand
      const x1 = Math.cos(angle) * radius;
      const y1 = (t - 0.5) * height;
      const z1 = Math.sin(angle) * radius;
      
      // Second strand (offset)
      const x2 = Math.cos(angle + Math.PI) * radius;
      const y2 = (t - 0.5) * height;
      const z2 = Math.sin(angle + Math.PI) * radius;
      
      // Create spheres for nodes
      const sphereGeometry = new THREE.SphereGeometry(0.3, 8, 8);
      const color1 = new THREE.Color().setHSL(0.5 + t * 0.2, 1, 0.5);
      const color2 = new THREE.Color().setHSL(0.7 + t * 0.2, 1, 0.5);
      
      const material1 = new THREE.MeshBasicMaterial({ color: color1 });
      const material2 = new THREE.MeshBasicMaterial({ color: color2 });
      
      const sphere1 = new THREE.Mesh(sphereGeometry, material1);
      sphere1.position.set(x1, y1, z1);
      
      const sphere2 = new THREE.Mesh(sphereGeometry, material2);
      sphere2.position.set(x2, y2, z2);
      
      helixGroup.add(sphere1);
      helixGroup.add(sphere2);
      
      // Connect every few points
      if (i % 5 === 0) {
        const lineGeometry = new THREE.BufferGeometry().setFromPoints([
          new THREE.Vector3(x1, y1, z1),
          new THREE.Vector3(x2, y2, z2)
        ]);
        const lineMaterial = new THREE.LineBasicMaterial({ 
          color: 0x00f0ff, 
          transparent: true, 
          opacity: 0.3 
        });
        const line = new THREE.Line(lineGeometry, lineMaterial);
        helixGroup.add(line);
      }
    }
    
    helixGroup.position.set(35, 0, -20);
    helixGroup.rotation.z = 0.3;
    this.scene.add(helixGroup);
  }

  private animate(): void {
    this.animationId = requestAnimationFrame(this.animate.bind(this));
    
    // Rotate particles
    if (this.particles) {
      this.particles.rotation.y += 0.0005;
      this.particles.rotation.x += 0.0002;
    }
    
    // Mouse interaction
    this.camera.position.x += (this.mouseX * 5 - this.camera.position.x) * 0.02;
    this.camera.position.y += (this.mouseY * 5 - this.camera.position.y) * 0.02;
    this.camera.lookAt(this.scene.position);
    
    // Rotate DNA helix
    this.scene.children.forEach(child => {
      if (child instanceof THREE.Group) {
        child.rotation.y += 0.005;
      }
    });
    
    this.renderer.render(this.scene, this.camera);
  }

  private onResize(): void {
    this.camera.aspect = window.innerWidth / window.innerHeight;
    this.camera.updateProjectionMatrix();
    this.renderer.setSize(window.innerWidth, window.innerHeight);
  }

  private initAnimations(): void {
    // Hero animations
    gsap.from('.hero-badge', {
      y: -30,
      opacity: 0,
      duration: 0.8,
      ease: 'power3.out'
    });
    
    gsap.from('.hero-title', {
      y: 50,
      opacity: 0,
      duration: 1,
      delay: 0.2,
      ease: 'power3.out'
    });
    
    gsap.from('.hero-subtitle', {
      y: 30,
      opacity: 0,
      duration: 0.8,
      delay: 0.4,
      ease: 'power3.out'
    });
    
    gsap.from('.hero-buttons', {
      y: 30,
      opacity: 0,
      duration: 0.8,
      delay: 0.6,
      ease: 'power3.out'
    });
    
    // Stats animations
    gsap.from('.stat-item', {
      y: 40,
      opacity: 0,
      duration: 0.6,
      stagger: 0.1,
      delay: 0.8,
      ease: 'power3.out'
    });
    
    // Feature cards animation on scroll
    gsap.from('.feature-card', {
      scrollTrigger: {
        trigger: '.features-grid',
        start: 'top 80%'
      },
      y: 60,
      opacity: 0,
      duration: 0.8,
      stagger: 0.15,
      ease: 'power3.out'
    });
  }
}
