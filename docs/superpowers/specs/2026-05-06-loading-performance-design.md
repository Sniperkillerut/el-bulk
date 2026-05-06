# High-Performance Loading System Design

## Goal
Improve the perceived performance and visual stability of product grids during initial load and filtering.

## Proposed Changes

### 1. Unified Shimmer Animation
Consolidate shimmer logic in `globals.css` to use a single, optimized animation.
- **Keyframes**: `shimmer` (defined in `globals.css`).
- **Optimization**: Speed up duration from `1.5s` to `0.8s` for a snappier feel.
- **Consistency**: Update `Skeleton.tsx` and `ProductGrid.tsx` to use the same animation name and timing.

### 2. New `ProductCardSkeleton` Component
Create a high-fidelity skeleton component that matches the `ProductCard` structure.
- **Location**: `frontend/src/components/skeletons/ProductCardSkeleton.tsx`
- **Structure**:
  - Card container with `card` class.
  - Aspect-ratio box for image.
  - Badge row placeholders.
  - Title and subtitle placeholders.
  - Footer placeholder with price and button shapes.
- **Benefits**: Eliminates layout shift when real data arrives.

### 3. Staggered Grid Entrance
Implement a subtle staggered entrance for products when they load.
- **Implementation**: Use a CSS utility or inline styles with `animation-delay` based on index.
- **Effect**: Products appear to "flow" in rather than popping in all at once.

### 4. ProductGrid Refactor
Replace inline skeletons in `ProductGrid.tsx` with the new `ProductCardSkeleton`.
- **Refactor**: Use a dedicated `LoadingState` sub-component or simply map the skeleton count.

## Verification Plan

### Automated Tests
- Verify `ProductCardSkeleton` renders correctly in isolation.
- Check that `ProductGrid` displays the correct number of skeletons when loading.

### Manual Verification
- Throttle network speed in DevTools to "Fast 3G" or "Slow 3G".
- Observe the transition from skeletons to real cards.
- Ensure no significant layout shifts occur.
- Verify shimmer animation speed feels appropriate.
