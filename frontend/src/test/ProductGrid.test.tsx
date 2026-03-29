import { render, screen, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import ProductGrid from '../components/ProductGrid'
import * as api from '@/lib/api'

// Mock the API using the alias
vi.mock('@/lib/api', () => ({
  fetchProducts: vi.fn().mockResolvedValue({ products: [], total: 0, facets: {} }),
  fetchCategories: vi.fn().mockResolvedValue([]),
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

const mockFacets = {
  condition: {},
  foil: {},
  treatment: {},
  rarity: {},
  category: {},
  language: {},
  color: {},
  collection: {}
}

import { act } from '@testing-library/react'

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
    // @ts-ignore
    vi.mocked(api.fetchProducts).mockResolvedValueOnce({
      products: null as any,
      total: 0,
      page: 1,
      page_size: 20,
      facets: mockFacets as any
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
      return {
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
            card_treatment: 'normal'
          }
        ] as any,
        total: 1,
        page: 1,
        page_size: 20,
        facets: mockFacets as any
      }
    })

    await act(async () => {
      render(<ProductGrid tcg="mtg" category="singles" title="MTG Singles" />)
    })

    await waitFor(() => {
      expect(screen.getByText('Black Lotus')).toBeInTheDocument()
    }, { timeout: 3000 })
  })
})
