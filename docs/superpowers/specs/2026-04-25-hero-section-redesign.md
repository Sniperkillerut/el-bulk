# Spec: El Bulk Hero & Navigation Redesign

## Goal
Redesign the home page hero section and main navigation to align with the "Clean Craft" brand identity: professional, humble, and specialized in bulk TCG cards.

## Design System

### Colors
- **Background**: `--bg-kraft` (Warm beige recycled paper texture: `#F4EDE4`)
- **Primary Ink**: `--ink-plum` (Muted Plum: `#5D3A4E`)
- **Secondary Ink**: `--ink-lavender` (Desaturated Lavender: `#938BB5` for subtle focus states)
- **Accent**: `--accent-rose` (Dusty Rose: `#B78494` for highlights)

### Typography
- **Headings (Brand)**: Classic Serif (e.g., EB Garamond)
- **Body & UI**: Clean Sans-Serif (e.g., Inter)
- **Labels (Optional)**: Typewriter Monospace (e.g., Cousine)

## Components

### 1. Unified Navigation Bar
- **Position**: Sticky top, single row.
- **Layout**:
  - **Left**: Minimalist "El Bulk" brand mark.
  - **Center**: Primary Navigation with thin line-art icons (consistent stroke weight):
    - 🎴 **Singles**
    - 📦 **Sealed**
    - 🧶 **Accessories**
    - ✨ **Exclusives**
    - 🔔 **News**
    - ✉️ **Contact**
    - 🎯 **Bounties**
  - **Right**: Utilities (👤 Log In, 🌐 Language, 🛒 Cart).
- **Style**: Transparent background on hero, transitioning to a solid `--bg-kraft` with a thin plum border-bottom on scroll.

### 2. Humble Hero Section
- **Layout**: Compact vertical arrangement.
- **Content**:
  - **Title**: "El Bulk" (Large, centered, Plum ink, reduced margin-top from Nav).
  - **Tagline**: "Grow your collection." (Subtle, centered directly below title).
  - **Search Bar**: 
    - Centered below tagline (minimal spacing).
    - Rectangular, thin plum border (`1px`).
    - Clean typography for placeholder.
    - Minimalist magnifying glass icon.
- **Background**: Subtle recycled paper texture overlay across the entire section.
- **Background**: Subtle recycled paper texture overlay across the entire section.

## Success Criteria
- No "vivid" or "neon" colors.
- Desktop-first balanced layout.
- Single navigation bar with no redundant links.
- "Handmade craft" feel achieved through texture and muted palette.
