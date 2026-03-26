import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock the fetch API globally
global.fetch = vi.fn()

// Mock Next.js navigation
vi.mock('next/navigation', () => ({
  notFound: vi.fn(),
  usePathname: vi.fn(() => '/'),
  useSearchParams: vi.fn(() => new URLSearchParams()),
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    back: vi.fn(),
  }),
}))

// Mock CartContext
vi.mock('@/lib/CartContext', () => ({
  useCart: () => ({
    cart: [],
    addItem: vi.fn(),
    removeItem: vi.fn(),
    clearCart: vi.fn(),
    total: 0,
  }),
  CartProvider: ({ children }: { children: React.ReactNode }) => children,
}))

// Mock ProductModalManager
vi.mock('@/components/ProductModalManager', () => ({
  openProductModal: vi.fn(),
}))
