const API = '/api';

export const ApiContract = {
  health: '/health',
  auth: {
    register: `${API}/register`,
    login: `${API}/login`,
    me: `${API}/me`
  },
  bootstrap: {
    get: `${API}/bootstrap`
  },
  notifications: {
    listMine: `${API}/notifications`
  },
  dashboard: {
    patientSummary: `${API}/patient/dashboard/summary`,
    doctorSummary: `${API}/doctor/dashboard/summary`
  },
  geoBooking: {
    geocode: `${API}/geocode`,
    hospitalsNear: `${API}/hospitals/near`,
    doctorsSearch: `${API}/doctors/search`,
    doctorSlots: (doctorId: string): string => `${API}/doctors/${doctorId}/slots`,
    createDoctorSlot: `${API}/doctor/slots`,
    deleteDoctorSlot: (slotId: string): string => `${API}/doctor/slots/${slotId}`,
    bookSlot: `${API}/appointments/book-slot`,
    doctors: `${API}/doctors`
  },
  appointments: {
    base: `${API}/appointments`,
    listPatient: `${API}/appointments/patient`,
    listDoctor: `${API}/appointments/doctor`,
    byId: (id: string): string => `${API}/appointments/${id}`,
    comments: (id: string): string => `${API}/appointments/${id}/comments`,
    activity: (id: string): string => `${API}/appointments/${id}/activity`,
    updateStatus: (id: string): string => `${API}/appointments/${id}/status`,
    patientCancel: (id: string): string => `${API}/patient/appointments/${id}/cancel`
  },
  documents: {
    base: `${API}/documents`,
    download: (id: string): string => `${API}/documents/${id}/download`
  },
  doctorDocuments: {
    base: `${API}/doctor/documents`,
    byId: (id: string): string => `${API}/doctor/documents/${id}`,
    summarize: (id: string): string => `${API}/doctor/documents/${id}/summarize`
  },
  patientFiles: {
    list: (patientId: string): string => `${API}/patients/${patientId}/files`,
    upload: (patientId: string): string => `${API}/patients/${patientId}/files`,
    download: (fileId: string): string => `${API}/files/${fileId}`
  },
  prescriptions: {
    listByPatient: (patientId: string): string => `${API}/patients/${patientId}/prescriptions`,
    createForPatient: (patientId: string): string => `${API}/patients/${patientId}/prescriptions`,
    revoke: (id: string): string => `${API}/prescriptions/${id}/revoke`
  },
  messaging: {
    unread: `${API}/messages/unread`,
    threads: `${API}/messages/threads`,
    threadMessages: (threadId: string): string => `${API}/messages/threads/${threadId}`,
    sendMessage: (threadId: string): string => `${API}/messages/threads/${threadId}/messages`,
    markRead: (threadId: string): string => `${API}/messages/threads/${threadId}/read`
  },
  admin: {
    auditLog: `${API}/admin/audit-log`,
    knowledgeDocs: `${API}/admin/knowledge-docs`,
    knowledgeDocById: (id: string): string => `${API}/admin/knowledge-docs/${id}`
  }
} as const;

