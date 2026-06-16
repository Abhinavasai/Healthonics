import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';

interface FeatureDetail {
  icon: string;
  title: string;
  tagline: string;
  overview: string;
  capabilities: { icon: string; title: string; description: string }[];
  howItWorks: { step: number; title: string; description: string }[];
  whoItsFor: string[];
  color: string;
}

@Component({
  selector: 'app-feature-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './feature-detail.component.html',
  styleUrl: './feature-detail.component.scss'
})
export class FeatureDetailComponent implements OnInit {
  feature: FeatureDetail | null = null;
  notFound = false;

  private readonly features: Record<string, FeatureDetail> = {
    'smart-health-monitoring': {
      icon: '🏥',
      title: 'Smart Health Monitoring',
      tagline: 'Real-time AI health tracking with predictive early warning systems',
      color: '#00f0ff',
      overview: `Healthonyx Smart Health Monitoring continuously analyzes your health data using advanced AI models to detect patterns, predict risks, and alert you — and your care team — before problems escalate. Instead of waiting for a scheduled visit to learn something is wrong, our platform gives you a live, intelligent picture of your health 24/7.`,
      capabilities: [
        { icon: '📊', title: 'Symptom Trend Analysis', description: 'AI tracks your reported symptoms over time, identifying patterns that may indicate developing conditions — even when individual symptoms seem minor.' },
        { icon: '⚠️', title: 'Early Warning Alerts', description: 'When data anomalies are detected, the system automatically generates critical escalation alerts for your assigned doctor to review immediately.' },
        { icon: '🤖', title: 'AI-Powered Symptom Checker', description: 'Describe how you feel in plain language. Our AI asks follow-up questions and gives structured guidance on urgency and next steps.' },
        { icon: '📈', title: 'Health Timeline', description: 'A full chronological view of all your health events — symptoms, lab results, prescriptions, and appointments — in a single scrollable timeline.' },
        { icon: '💡', title: 'Predictive Insights', description: 'Machine learning models trained on clinical data surface personalized insights: which medications have interaction risks, which symptoms cluster together, and when to seek care.' },
        { icon: '🔔', title: 'Smart Notifications', description: 'Medication reminders, appointment alerts, and lab result notifications delivered at the right time through your preferred channel.' }
      ],
      howItWorks: [
        { step: 1, title: 'You log symptoms or check in', description: 'Use the Symptom Checker to describe how you feel, or respond to a scheduled check-in notification.' },
        { step: 2, title: 'AI analyzes the data', description: 'Our AI processes your entry against your full health history, current prescriptions, and clinical knowledge base.' },
        { step: 3, title: 'Insights are generated', description: 'The system produces a structured assessment — urgency level, possible explanations, and recommended actions.' },
        { step: 4, title: 'Your doctor is notified if needed', description: 'High-urgency findings trigger an immediate alert to your care team. Low-urgency findings are summarized in your next appointment.' }
      ],
      whoItsFor: ['Patients managing chronic conditions', 'Caregivers monitoring family members', 'Healthcare providers wanting proactive patient oversight', 'Anyone who wants to stay ahead of their health']
    },

    'digital-medical-records': {
      icon: '📋',
      title: 'Digital Medical Records',
      tagline: 'Your complete health history — secure, encrypted, and always accessible',
      color: '#7b61ff',
      overview: `Healthonyx replaces fragmented, paper-based records with a unified, encrypted digital health record. Every appointment note, prescription, lab result, uploaded document, and clinical summary lives in one place — accessible to you and your authorized care team instantly, from anywhere.`,
      capabilities: [
        { icon: '📁', title: 'Centralized Document Storage', description: 'Upload PDFs, images, and reports from any device. All files are AES-256 encrypted at rest and in transit.' },
        { icon: '🧪', title: 'Lab Result Tracking', description: 'Lab reports are stored and indexed. Our AI can provide a plain-language interpretation of complex lab values so you understand what they mean.' },
        { icon: '💊', title: 'Prescription History', description: 'Every prescription ever issued to you — active, expired, or revoked — is stored with dosage, frequency, prescribing doctor, and notes.' },
        { icon: '📤', title: 'FHIR Export', description: 'Export your complete health record in HL7 FHIR R4 format — the international healthcare interoperability standard — to share with any provider worldwide.' },
        { icon: '📄', title: 'Health Summary PDF', description: 'Generate a concise health summary document at any time, covering your conditions, medications, allergies, and recent visits.' },
        { icon: '🔐', title: 'Access Control', description: 'You control who sees your records. Grant or revoke access to specific providers at any time, with a full audit trail of every access event.' }
      ],
      howItWorks: [
        { step: 1, title: 'Records are created automatically', description: 'Appointments, prescriptions, and notes are logged automatically by your care team during consultations.' },
        { step: 2, title: 'You can upload your own', description: 'Upload any health documents — old test results, referral letters, imaging reports — directly from the patient dashboard.' },
        { step: 3, title: 'AI organizes and interprets', description: 'Documents are classified, indexed, and made searchable. AI-generated summaries help you understand complex medical language.' },
        { step: 4, title: 'Share securely when needed', description: 'Share specific records or your full FHIR bundle with any provider using a secure, time-limited link or direct export.' }
      ],
      whoItsFor: ['Patients with complex medical histories', 'People who see multiple specialists', 'Individuals managing chronic conditions', 'Healthcare providers needing a complete patient picture']
    },

    'doctor-portal': {
      icon: '👨‍⚕️',
      title: 'Doctor Portal',
      tagline: 'A comprehensive command center for modern healthcare providers',
      color: '#00c896',
      overview: `The Healthonyx Doctor Portal gives physicians, specialists, and nurse practitioners a powerful, real-time workspace to manage their entire patient panel. From reviewing appointment queues to writing AI-assisted clinical notes, the portal is designed to reduce administrative overhead and let doctors focus on what matters — patient care.`,
      capabilities: [
        { icon: '🗂️', title: 'Patient Panel', description: 'See all your patients at a glance — last appointment date, active prescriptions, and pending action items for each one.' },
        { icon: '📅', title: 'Appointment Management', description: 'View, approve, reschedule, or cancel appointments. Set video consultation links for teleconsult sessions. Add clinical notes and status updates.' },
        { icon: '🚪', title: 'Waiting Room Queue', description: 'Real-time view of today\'s patients sorted by check-in time. See how long each patient has been waiting and manage flow efficiently.' },
        { icon: '📝', title: 'AI Note Assistant', description: 'Dictate or type rough notes and let AI structure them into professional SOAP-format clinical notes, saving significant documentation time.' },
        { icon: '💊', title: 'Prescription Writing', description: 'Issue digital prescriptions with built-in drug interaction checking. Prescriptions are instantly available to the patient and logged to their record.' },
        { icon: '🔁', title: 'Refill Requests', description: 'Review and approve or deny patient refill requests in one click, with the ability to add notes explaining the decision.' },
        { icon: '📢', title: 'Bulk Messaging', description: 'Send a message to all your patients at once — ideal for clinic announcements, health advisories, or follow-up campaigns.' },
        { icon: '⭐', title: 'Patient Ratings', description: 'View your aggregated patient satisfaction scores and individual feedback comments to continuously improve care quality.' }
      ],
      howItWorks: [
        { step: 1, title: 'Log in to your portal', description: 'Access your personalized dashboard showing today\'s schedule, critical escalations, and unread messages at a glance.' },
        { step: 2, title: 'Manage your day', description: 'Work through appointments, check the waiting room queue, and respond to refill requests — all from a single screen.' },
        { step: 3, title: 'Document with AI assistance', description: 'Use the AI Note Assistant to draft clinical notes faster. The AI understands medical context and formats notes to clinical standards.' },
        { step: 4, title: 'Stay connected with patients', description: 'Secure messaging, teleconsult video links, and push notifications keep you connected with patients between visits.' }
      ],
      whoItsFor: ['General practitioners', 'Specialists and consultants', 'Nurse practitioners', 'Clinic administrators managing provider schedules']
    },

    'smart-appointments': {
      icon: '📅',
      title: 'Smart Appointments',
      tagline: 'AI-assisted scheduling with automated reminders and teleconsult support',
      color: '#ff9500',
      overview: `Booking, managing, and attending appointments should be effortless. Healthonyx Smart Appointments uses AI to recommend the best available slots based on your history and urgency, automates reminders so no one misses a visit, and supports full teleconsultations with built-in video link integration.`,
      capabilities: [
        { icon: '🤖', title: 'AI Slot Recommendations', description: 'Tell us when you\'re available and what you need. AI ranks open slots by relevance — considering doctor specialty, your history, and urgency level.' },
        { icon: '🔔', title: 'Automated Reminders', description: 'Patients receive automatic reminders 24 hours and 1 hour before their appointment, reducing no-shows significantly.' },
        { icon: '📹', title: 'Teleconsult Video Links', description: 'Doctors can attach a Jitsi Meet (or any HTTPS video link) to an appointment. Patients see the link on their appointment detail page and can join with one tap.' },
        { icon: '✅', title: 'Digital Check-In', description: 'Patients check in from their phone before arriving. The doctor\'s waiting room queue updates in real time.' },
        { icon: '📋', title: 'Pre-Visit Questionnaire', description: 'Before each appointment, patients complete a customizable questionnaire. Answers are available to the doctor before the visit starts.' },
        { icon: '📊', title: 'Workload Heatmap', description: 'Doctors can view a heatmap of their scheduled appointments to identify busy periods and optimize availability.' }
      ],
      howItWorks: [
        { step: 1, title: 'Browse available doctors', description: 'Search by specialty, location, or availability. Filter by teleconsult-capable providers if you prefer a virtual visit.' },
        { step: 2, title: 'Book your slot', description: 'Select a time that works for you. The system confirms instantly and sends a calendar invite.' },
        { step: 3, title: 'Prepare for your visit', description: 'Complete the pre-visit questionnaire and check in digitally on the day of your appointment.' },
        { step: 4, title: 'Attend and follow up', description: 'Join via video link for teleconsults or arrive in person. After the visit, rate your experience and receive any follow-up instructions.' }
      ],
      whoItsFor: ['Patients with busy schedules', 'People in remote areas who need teleconsult', 'Doctors wanting to reduce no-shows', 'Clinics optimizing scheduling efficiency']
    },

    'prescription-management': {
      icon: '💊',
      title: 'Prescription Management',
      tagline: 'Digital prescriptions with interaction checking and refill automation',
      color: '#ff3b6f',
      overview: `Healthonyx digitizes the entire prescription lifecycle — from a doctor writing a new prescription to a patient requesting a refill. Built-in drug interaction checking catches dangerous combinations before they reach the patient, and the full prescription history is always available to the care team.`,
      capabilities: [
        { icon: '✍️', title: 'Digital Prescription Writing', description: 'Doctors issue fully digital prescriptions specifying medication, dosage, frequency, duration, and instructions — no paper, no handwriting to decipher.' },
        { icon: '⚠️', title: 'Drug Interaction Checking', description: 'AI automatically checks new prescriptions against the patient\'s current medication list and flags potential interactions before the prescription is finalized.' },
        { icon: '📄', title: 'Prescription PDF', description: 'Patients can download a formatted PDF of any prescription to present at a pharmacy or share with another provider.' },
        { icon: '🔁', title: 'Refill Requests', description: 'Patients request refills with a single tap. Doctors review and approve or deny with an optional note. The patient is notified immediately.' },
        { icon: '📅', title: 'Expiry Tracking', description: 'The system tracks prescription expiry dates and notifies patients before their medication runs out, preventing dangerous gaps in treatment.' },
        { icon: '📊', title: 'Adherence Scoring', description: 'AI analyzes refill patterns and symptom reports to generate a medication adherence score — helping doctors identify patients who may need additional support.' }
      ],
      howItWorks: [
        { step: 1, title: 'Doctor writes the prescription', description: 'From the appointment detail page, the doctor fills in the prescription form. Drug interaction AI runs automatically.' },
        { step: 2, title: 'Patient receives instant notification', description: 'The patient is notified immediately and can view the full prescription in their records.' },
        { step: 3, title: 'Patient requests a refill when needed', description: 'Before the prescription expires, the patient submits a one-tap refill request with an optional note to the doctor.' },
        { step: 4, title: 'Doctor reviews and responds', description: 'The doctor approves or denies the refill from the Doctor Portal. The patient is notified of the outcome.' }
      ],
      whoItsFor: ['Patients on long-term medications', 'Doctors managing complex medication regimens', 'Clinics reducing prescription errors', 'Pharmacies needing clear, legible digital prescriptions']
    },

    'enterprise-security': {
      icon: '🔒',
      title: 'Enterprise Security',
      tagline: 'HIPAA-compliant infrastructure with end-to-end encryption and full audit trails',
      color: '#7b61ff',
      overview: `Healthcare data is among the most sensitive information in existence. Healthonyx was architected from the ground up with security as a first-class requirement — not an afterthought. Every byte of patient data is encrypted, every access is logged, and every system component is built to meet or exceed HIPAA requirements.`,
      capabilities: [
        { icon: '🔐', title: 'Envelope Encryption', description: 'All sensitive data (messages, documents, clinical notes) is encrypted using AES-256 with per-record key wrapping. Even a database breach exposes only ciphertext.' },
        { icon: '📋', title: 'Full Audit Trail', description: 'Every action in the system — who logged in, who viewed which record, who changed what — is written to an immutable audit log.' },
        { icon: '🔑', title: 'JWT Authentication', description: 'Secure short-lived JSON Web Tokens for authentication. Tokens expire automatically and can be invalidated instantly on logout.' },
        { icon: '🛡️', title: 'Role-Based Access Control', description: 'Strict role separation between patients, doctors, and admins. Users can only access data they are explicitly authorized to see.' },
        { icon: '📊', title: 'SLO Monitoring', description: 'An admin-facing SLO dashboard tracks API latency percentiles (P50/P95/P99), error rates, and uptime — enabling proactive identification of system issues.' },
        { icon: '🔍', title: 'Security Review Tooling', description: 'Built-in security review processes and automated checks run continuously to identify potential vulnerabilities before they become incidents.' }
      ],
      howItWorks: [
        { step: 1, title: 'Data is encrypted at entry', description: 'The moment sensitive data enters the system, it is encrypted before being written to the database. Decryption only happens when an authorized user requests it.' },
        { step: 2, title: 'Every access is authenticated and authorized', description: 'JWT tokens are validated on every API call. Role checks ensure users can only reach their permitted endpoints.' },
        { step: 3, title: 'All actions are logged', description: 'The audit log records every meaningful action with timestamp, user identity, IP address, and outcome.' },
        { step: 4, title: 'Administrators monitor in real time', description: 'The SLO dashboard gives admins live visibility into system health, error rates, and performance so issues are caught before they affect patients.' }
      ],
      whoItsFor: ['Healthcare organizations subject to HIPAA', 'IT administrators managing clinical systems', 'Compliance officers needing audit evidence', 'Patients who want assurance their data is protected']
    }
  };

  constructor(private route: ActivatedRoute) {}

  ngOnInit(): void {
    this.route.paramMap.subscribe(params => {
      const slug = params.get('slug') ?? '';
      this.feature = this.features[slug] ?? null;
      this.notFound = !this.feature;
    });
  }
}
