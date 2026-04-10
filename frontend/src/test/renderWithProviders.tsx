import React, { ReactElement } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { LanguageProvider } from '@/context/LanguageContext'
import { UIProvider } from '@/context/UIContext'
import { CartProvider } from '@/lib/CartContext'
import { UserProvider } from '@/context/UserContext'

interface AllTheProvidersProps {
  children: React.ReactNode
}

const AllTheProviders = ({ children }: AllTheProvidersProps) => {
  return (
    <LanguageProvider>
      <UserProvider>
        <UIProvider>
          <CartProvider>
            {children}
          </CartProvider>
        </UIProvider>
      </UserProvider>
    </LanguageProvider>
  )
}

const renderWithProviders = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, { wrapper: AllTheProviders, ...options })

export * from '@testing-library/react'
export { renderWithProviders as render }
