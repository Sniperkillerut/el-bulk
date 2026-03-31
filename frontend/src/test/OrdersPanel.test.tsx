import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, expect, it, describe, beforeEach } from 'vitest';
import OrdersPanel from '../components/admin/OrdersPanel';
import * as api from '@/lib/api';

// Mock the API module
vi.mock('@/lib/api', () => ({
  adminFetchOrders: vi.fn(),
  adminFetchOrderDetail: vi.fn(),
  adminUpdateOrder: vi.fn(),
  adminCompleteOrder: vi.fn(),
}));

const mockOrders = [
  {
    id: 'o1',
    order_number: 'EB-1',
    customer_id: 'c1',
    customer_name: 'John Doe',
    status: 'pending' as const,
    total_cop: 1000,
    created_at: new Date().toISOString(),
    item_count: 1,
    payment_method: 'whatsapp',
  },
];

const mockOrderDetail = {
  order: mockOrders[0],
  customer: {
    id: 'c1',
    first_name: 'John',
    last_name: 'Doe',
    phone: '123',
    email: 'john@doe.com',
    created_at: new Date().toISOString(),
  },
  items: [
    {
      id: 'oi1',
      order_id: 'o1',
      product_id: 'p1',
      product_name: 'Black Lotus',
      quantity: 1,
      unit_price_cop: 1000,
      stock: 5,
      stored_in: [{ stored_in_id: 's1', name: 'Box 1', quantity: 5 }],
    },
  ],
};

describe('OrdersPanel', () => {
  const token = 'test-token';

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.adminFetchOrders).mockResolvedValue({
      orders: mockOrders,
      total: 1,
      page: 1,
      page_size: 20,
    });
    vi.mocked(api.adminFetchOrderDetail).mockResolvedValue(mockOrderDetail);
  });

  it('renders leading state initially', async () => {
    render(<OrdersPanel token={token} />);
    expect(screen.getByText(/Cargando\.\.\./i)).toBeInTheDocument();
  });

  it('renders orders after loading', async () => {
    render(<OrdersPanel token={token} />);
    
    await waitFor(() => {
      expect(screen.getByText('EB-1')).toBeInTheDocument();
      expect(screen.getByText('John Doe')).toBeInTheDocument();
    });
  });

  it('shows order details when an order is clicked', async () => {
    render(<OrdersPanel token={token} />);
    
    await waitFor(() => {
      const orderItem = screen.getByText('EB-1');
      fireEvent.click(orderItem);
    });

    await waitFor(() => {
      expect(api.adminFetchOrderDetail).toHaveBeenCalledWith(token, 'o1');
      expect(screen.getByText(/Black Lotus/i)).toBeInTheDocument();
      expect(screen.getByText('John Doe →')).toBeInTheDocument();
    });
  });

  it('filters orders by search input', async () => {
    render(<OrdersPanel token={token} />);
    
    const searchInput = screen.getByPlaceholderText(/Buscar por # orden/i);
    fireEvent.change(searchInput, { target: { value: 'EB-2' } });

    // Wait for debounce (300ms)
    await waitFor(() => {
      expect(api.adminFetchOrders).toHaveBeenCalledWith(token, expect.objectContaining({
        search: 'EB-2'
      }));
    }, { timeout: 1000 });
  });

  it('filters orders by status', async () => {
    render(<OrdersPanel token={token} />);
    
    const pendingButton = screen.getByRole('button', { name: /Pendiente/i });
    fireEvent.click(pendingButton);

    await waitFor(() => {
      expect(api.adminFetchOrders).toHaveBeenCalledWith(token, expect.objectContaining({
        status: 'pending'
      }));
    });
  });
});
