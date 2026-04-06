import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import HomeSearchBar from '../components/HomeSearchBar'
import { useLanguage } from "@/context/LanguageContext"
vi.mock("@/context/LanguageContext", () => ({
  useLanguage: vi.fn(),
}))
import * as api from '@/lib/api'

// Mock the API using the alias
vi.mock('@/lib/api', () => ({
  fetchProducts: vi.fn().mockResolvedValue({ products: [], total: 0, facets: {} }),
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

describe('HomeSearchBar', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useLanguage).mockReturnValue({
      t: (key: string) => key,
      locale: "en",
      setLocale: vi.fn(),
      isLoading: false,
      availableLocales: ["en", "es"],
    })
  })

  it('renders correctly', async () => {
    await act(async () => {
      render(<HomeSearchBar />)
    })
    expect(screen.getByPlaceholderText(/search for cards/i)).toBeInTheDocument()
  })

  it('handles empty search results gracefully', async () => {
    vi.mocked(api.fetchProducts).mockImplementation(async () => {
      return {
        products: [],
        total: 0,
        page: 1,
        page_size: 10,
        facets: mockFacets as any,
        query_time_ms: 0
      }
    })

    await act(async () => {
      render(<HomeSearchBar />)
    })
    const input = screen.getByPlaceholderText(/search for cards/i)
    
    // Type something to trigger search
    await act(async () => {
      fireEvent.change(input, { target: { value: 'black lotus' } })
    })

    // Wait for debounce and API call
    await waitFor(() => {
      expect(api.fetchProducts).toHaveBeenCalled()
    }, { timeout: 2000 })

    // Use a more robust matcher
    await waitFor(() => {
      expect(screen.getByText((content) => content.includes('No products found'))).toBeInTheDocument()
    }, { timeout: 2000 })
  })

  it('displays search results when found', async () => {
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
            card_treatment: 'normal',
            language: 'en',
            is_legendary: false,
            is_historic: false,
            is_land: false,
            is_basic_land: false,
            full_art: false,
            textless: false,
            created_at: '',
            updated_at: ''
          }
        ] as any,
        total: 1,
        page: 1,
        page_size: 10,
        facets: mockFacets as any,
        query_time_ms: 0
      }
    })

    await act(async () => {
      render(<HomeSearchBar />)
    })
    const input = screen.getByPlaceholderText(/search for cards/i)
    
    await act(async () => {
      fireEvent.change(input, { target: { value: 'black lotus' } })
    })

    await waitFor(() => {
      expect(screen.getByText('Black Lotus')).toBeInTheDocument()
    }, { timeout: 2000 })
  })
})
