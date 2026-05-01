import { render, screen, waitFor } from './renderWithProviders'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import ProductGrid from '../components/ProductGrid'
import * as api from '@/lib/api'
import { Product, ProductListResponse, Facets } from '@/lib/types'

// Mock the API using the alias
vi.mock('@/lib/api', () => ({
  fetchProducts: vi.fn().mockResolvedValue({ products: [], total: 0, facets: {} }),
  fetchCategories: vi.fn().mockResolvedValue([]),
  fetchTCGs: vi.fn().mockResolvedValue([]),
  fetchTranslations: vi.fn().mockResolvedValue({}),
  fetchSettings: vi.fn().mockResolvedValue({}),
  fetchFacets: vi.fn(),
  getProxyImageUrl: vi.fn((url) => url),
}))

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => '',
}))

const mockFacets: Facets = {
  condition: {},
  foil: {},
  treatment: {},
  rarity: {},
  language: {},
  color: {},
  collection: {}
}

import { act } from './renderWithProviders'

describe('ProductGrid', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.fetchCategories).mockResolvedValue([])
  })

  it('renders title and loading state', async () => {
    await act(async () => {
      render(<ProductGrid tcg="mtg" category="singles" title="MTG Singles" />)
    })
    expect(screen.getByText('MTG Singles')).toBeInTheDocument()
  })

  it('handles null products gracefully and shows NO RESULTS', async () => {
    vi.mocked(api.fetchProducts).mockResolvedValueOnce({
      products: null as unknown as Product[],
      total: 0,
      page: 1,
      page_size: 20,
      facets: mockFacets,
      query_time_ms: 0
    })

    await act(async () => {
      render(<ProductGrid tcg="mtg" category="singles" title="MTG Singles" />)
    })

    await waitFor(() => {
      expect(screen.getByText(/NO RESULTS/i)).toBeInTheDocument()
    })
  })

  it('displays products when found', async () => {
    // Need to use mockImplementation instead of mockResolvedValueOnce because useSWR cache may re-fetch
    vi.mocked(api.fetchProducts).mockImplementation(async () => {
      const resp: ProductListResponse = {
        products: [
          {
            id: '1',
            name: 'Black Lotus',
            tcg: 'mtg',
            price: 100000,
            image_url: '',
            category: 'singles',
            stock: 1,
            condition: 'NM',
            foil_treatment: 'non_foil',
            card_treatment: 'normal',
            language: 'en',
            price_source: 'tcgplayer',
            is_legendary: false,
            is_historic: false,
            is_land: false,
            is_basic_land: false,
            full_art: false,
            textless: false,
            created_at: '',
            updated_at: ''
          }
        ],
        total: 1,
        page: 1,
        page_size: 20,
        facets: mockFacets,
        query_time_ms: 0
      }
      return resp
    })

    await act(async () => {
      render(<ProductGrid tcg="mtg" category="singles" title="MTG Singles" />)
    })

    await waitFor(() => {
      expect(screen.getByText('Black Lotus')).toBeInTheDocument()
    }, { timeout: 3000 })
  })
})
