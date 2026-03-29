import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect } from 'vitest'
import CardImage from '../components/CardImage'

describe('CardImage', () => {
  it('renders placeholder when no imageUrl is provided', () => {
    render(<CardImage imageUrl={null} name="Test Card" tcg="mtg" />)
    expect(screen.getByText('⚔️')).toBeInTheDocument()
    expect(screen.getByText('TEST CARD')).toBeInTheDocument()
  })

  it('renders image when imageUrl is provided', () => {
    render(<CardImage imageUrl="https://example.com/image.jpg" name="Test Card" tcg="mtg" />)
    const img = screen.getByRole('img')
    expect(img).toHaveAttribute('src', expect.stringContaining('image.jpg'))
  })
})
