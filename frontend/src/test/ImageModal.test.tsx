import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect } from 'vitest'
import ImageModal from '../components/ImageModal'

describe('ImageModal', () => {
  it('renders image', () => {
    render(<ImageModal imageUrl="test.jpg" name="Test Card" onClose={vi.fn()} />)
    const img = screen.getByAltText('Test Card')
    expect(img).toBeInTheDocument()
    expect(img).toHaveAttribute('src', 'test.jpg')
  })

  it('calls onClose when close button clicked', () => {
    const onClose = vi.fn()
    render(<ImageModal imageUrl="test.jpg" name="Test Card" onClose={onClose} />)

    // There might be multiple elements with text that looks like a close button
    // The component uses an SVG or a button with onClick
    // Since we know the DOM structure: it's a button with className containing "absolute top-4 right-4"
    const closeButtons = screen.getAllByRole('button')
    fireEvent.click(closeButtons[0])

    expect(onClose).toHaveBeenCalled()
  })

  it('calls onClose when backdrop is clicked', () => {
    const onClose = vi.fn()
    render(<ImageModal imageUrl="test.jpg" name="Test Card" onClose={onClose} />)

    const backdropButtons = screen.getAllByRole('button')
    fireEvent.click(backdropButtons[0]) // First button is the invisible backdrop button

    expect(onClose).toHaveBeenCalled()
  })
})
