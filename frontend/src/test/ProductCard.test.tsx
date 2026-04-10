import { render, screen, fireEvent } from './renderWithProviders'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import ProductCard from '../components/ProductCard'
import { useCart } from '@/lib/CartContext'

const mockProduct = {
  id: 'test-1',
  name: 'Test Product',
  tcg: 'mtg',
  price: 1000,
  image_url: '/test.jpg',
  category: 'singles',
  stock: 5,
  condition: 'NM',
  foil_treatment: 'foil',
  card_treatment: 'full_art'
}

describe('ProductCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders product details correctly', () => {
    render(<ProductCard product={mockProduct} />)
    expect(screen.getByText('Test Product')).toBeInTheDocument()
    expect(screen.getByText('NM')).toBeInTheDocument()
    expect(screen.getByText('✦ Foil')).toBeInTheDocument()
    expect(screen.getByText('Full Art (Regular)')).toBeInTheDocument()
    expect(screen.getByText(/\$1,000/)).toBeInTheDocument()
    expect(screen.getByText(/COP/)).toBeInTheDocument()
  })

  it('handles "out of stock" state', () => {
    render(<ProductCard product={{ ...mockProduct, stock: 0 }} />)
    expect(screen.getByText('SOLD OUT')).toBeInTheDocument()
    const addButton = screen.getByText('SOLD OUT')
    expect(addButton).toBeDisabled()
  })

  it('calls addItem when Add button is clicked', () => {
    const { addItem } = useCart()
    render(<ProductCard product={mockProduct} />)
    const addButton = screen.getByText('ADD')
    fireEvent.click(addButton)
    expect(addItem).toHaveBeenCalledWith(mockProduct)
  })

  it('renders correctly when product has textless and alternate categories', () => {
    const product = {
      ...mockProduct,
      textless: true,
      categories: [{ id: 'cat1', name: 'Commander Staples', slug: 'commander', is_active: true }]
    }
    const { container } = render(<ProductCard product={product as any} />)
    expect(screen.getByText('TEXTLESS')).toBeInTheDocument()
    expect(screen.getByText(/Commander Staples/i)).toBeInTheDocument()
  })

  it('renders without foil badge when non_foil', () => {
    render(<ProductCard product={{ ...mockProduct, foil_treatment: 'non_foil' }} />)
    expect(screen.queryByText('✦ Foil')).not.toBeInTheDocument()
  })
})
