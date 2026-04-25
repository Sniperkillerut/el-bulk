# Hero & Navigation Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the home page hero and navigation to a "Clean Craft" aesthetic with a unified, icon-rich navbar and a compact hero layout.

**Architecture:** 
1.  **Design Tokens**: Update `globals.css` with `--ink-plum`, `--bg-kraft`, and other craft-themed variables.
2.  **Shared Icons**: Create `frontend/src/components/ui/LineIcons.tsx` for consistent, minimalist SVG icons.
3.  **Unified Navbar**: Refactor `Navbar.tsx` to a single-row layout containing all requested links with icons.
4.  **Compact Hero**: Redesign `HeroSection.tsx` to reduce vertical whitespace and apply the paper-texture look.

**Tech Stack:** React (Next.js), Tailwind CSS, Vanilla CSS.

---

### Task 1: Design Tokens & Base Styles
**Files:**
- Modify: `frontend/src/app/globals.css`

- [ ] **Step 1: Define craft color variables and paper texture.**
```css
:root {
  /* Clean Craft Palette */
  --bg-kraft: #F4EDE4;
  --ink-plum: #5D3A4E;
  --ink-lavender: #938BB5;
  --accent-rose: #B78494;
  --border-plum: rgba(93, 58, 78, 0.2);
  
  /* Update existing functional variables if applicable */
  --accent-primary: var(--ink-plum);
  --text-main: var(--ink-plum);
}

.paper-texture {
  position: relative;
}

.paper-texture::before {
  content: "";
  position: absolute;
  inset: 0;
  opacity: 0.05;
  pointer-events: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='400'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='3' stitchTiles='stitch'/%3E%3CfeColorMatrix type='saturate' values='0'/%3E%3C/filter%3E%3Crect width='400' height='400' filter='url(%23noise)' opacity='1'/%3E%3C/svg%3E");
  z-index: 1;
}
```

- [ ] **Step 2: Commit base styles.**

---

### Task 2: Minimalist Icon Library
**Files:**
- Create: `frontend/src/components/ui/LineIcons.tsx`

- [ ] **Step 1: Implement shared minimalist icons.**
```tsx
export const LineIcons = {
  Singles: () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <rect x="5" y="3" width="14" height="18" rx="2" ry="2" />
      <path d="M9 7h6" />
    </svg>
  ),
  Sealed: () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z" />
      <path d="m3.3 7 8.7 5 8.7-5" /><path d="M12 22V12" />
    </svg>
  ),
  Accessories: () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" /><path d="M12 8v8" /><path d="M8 12h8" />
    </svg>
  ),
  Exclusives: () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 2v20" /><path d="m4.93 4.93 14.14 14.14" /><path d="M2 12h20" /><path d="m4.93 19.07 14.14-14.14" />
    </svg>
  ),
  News: () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4 4h16v16H4z" /><path d="M8 8h8" /><path d="M8 12h8" /><path d="M8 16h4" />
    </svg>
  ),
  Contact: () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="4" width="20" height="16" rx="2" />
      <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
    </svg>
  ),
  Bounties: () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" /><circle cx="12" cy="12" r="6" /><circle cx="12" cy="12" r="2" />
    </svg>
  )
};
```

- [ ] **Step 2: Commit icons.**

---

### Task 3: Unified Navbar Overhaul
**Files:**
- Modify: `frontend/src/components/Navbar.tsx`

- [ ] **Step 1: Import icons and update nav layout to single row.**
- [ ] **Step 2: Implement the condensed 7-item menu with icons.**
- [ ] **Step 3: Update mobile menu to reflect the same icons.**
- [ ] **Step 4: Commit.**

---

### Task 4: Compact Hero Section
**Files:**
- Modify: `frontend/src/components/HeroSection.tsx`

- [ ] **Step 1: Apply paper texture and craft palette.**
- [ ] **Step 2: Tighten vertical spacing between Nav, Brand Title, and Search.**
- [ ] **Step 3: Update Brand Title to use the humble "Grow your collection" tagline.**
- [ ] **Step 4: Commit.**

---

### Task 5: Search Bar Polishing
**Files:**
- Modify: `frontend/src/components/HomeSearchBar.tsx`

- [ ] **Step 1: Simplify border and typography (Plum 1px border).**
- [ ] **Step 2: Ensure it aligns with the "Clean Craft" look.**
- [ ] **Step 3: Commit.**
