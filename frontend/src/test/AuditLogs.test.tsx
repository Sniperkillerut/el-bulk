import { render, screen, waitFor, fireEvent } from './renderWithProviders'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import AuditLogsPage from '../app/admin/audit-logs/page'
import * as api from '@/lib/api'

// Mock the API
vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual('@/lib/api') as any
  return {
    ...actual,
    adminFetchAuditLogs: vi.fn(),
    fetchTranslations: vi.fn().mockResolvedValue({}),
    fetchSettings: vi.fn().mockResolvedValue({}),
    getProxyImageUrl: vi.fn((url) => url),
  }
})

// Mock useAdmin hook
vi.mock('@/hooks/useAdmin', () => ({
  useAdmin: () => ({
    token: 'mock-token',
    settings: {},
    loading: false,
  }),
}))

const mockLogs = {
  logs: [
    {
      id: '1',
      admin_id: 'admin-id-1',
      admin_username: 'admin1',
      action: 'create',
      resource_type: 'product',
      resource_id: 'prod-123',
      details: { name: 'Test' },
      created_at: new Date().toISOString(),
    },
  ],
  total: 1,
  page: 1,
  page_size: 20,
}

describe('AuditLogsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.adminFetchAuditLogs).mockResolvedValue(mockLogs)
  })

  it('renders correctly and fetches logs', async () => {
    render(<AuditLogsPage />)

    expect(screen.getByText(/ACTION LOG/i)).toBeInTheDocument()
    expect(screen.getByText('Synchronizing ledger...')).toBeInTheDocument()

    await waitFor(() => {
      expect(api.adminFetchAuditLogs).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(screen.getByText('admin1')).toBeInTheDocument()
      expect(screen.getByText('product')).toBeInTheDocument()
      expect(screen.getByText(/ID: prod-123/i)).toBeInTheDocument()
    })
  })

  it('filters logs when action selection changes', async () => {
    render(<AuditLogsPage />)

    // Since there are two selects, let's be more specific if possible or just use the first one
    const selects = screen.getAllByRole('combobox')
    // First select is Action
    vi.mocked(api.adminFetchAuditLogs).mockClear()
    // Trigger change
    const actionSelectElement = selects[0] as HTMLSelectElement
    vi.mocked(api.adminFetchAuditLogs).mockResolvedValue(mockLogs)

    fireEvent.change(actionSelectElement, { target: { value: 'CREATE' } })

    await waitFor(() => {
      expect(api.adminFetchAuditLogs).toHaveBeenCalledWith(expect.objectContaining({
        action: 'CREATE'
      }))
    })
  })
})
