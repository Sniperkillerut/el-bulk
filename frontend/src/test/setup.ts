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

// Mock next/link
vi.mock('next/link', () => {
  return {
    default: ({ children, href }: { children: any; href: string }) => {
      return children;
    },
  };
});

// Mock swr globally to properly handle promises and re-renders
vi.mock('swr', () => {
  let cache = new Map();
  return {
    default: vi.fn((key, fetcher) => {
      const [data, setData] = require('react').useState(cache.get(JSON.stringify(key)));
      const [error, setError] = require('react').useState(null);
      const [isLoading, setIsLoading] = require('react').useState(!cache.has(JSON.stringify(key)));

      require('react').useEffect(() => {
        if (typeof fetcher === 'function') {
          setIsLoading(true);
          Promise.resolve(fetcher(key)).then((result) => {
            cache.set(JSON.stringify(key), result);
            setData(result);
            setIsLoading(false);
          }).catch((err) => {
            setError(err);
            setIsLoading(false);
          });
        }
      }, [JSON.stringify(key)]); // stringify key to handle object dependencies like fetcherArgs

      return { data, error, isLoading };
    }),
    useSWRConfig: () => ({ cache: new Map() }), // Mock cache if needed
  };
});

// Mock CartContext
vi.mock('@/lib/CartContext', () => ({
  useCart: () => ({
    items: [],
    totalItems: 0,
    totalPrice: 0,
    addItem: vi.fn(),
    removeItem: vi.fn(),
    updateQty: vi.fn(),
    clearCart: vi.fn(),
    isOpen: false,
    openCart: vi.fn(),
    closeCart: vi.fn(),
  }),
  CartProvider: ({ children }: { children: React.ReactNode }) => children,
}))

// Mock ProductModalManager
vi.mock('@/components/ProductModalManager', () => ({
  openProductModal: vi.fn(),
}))
