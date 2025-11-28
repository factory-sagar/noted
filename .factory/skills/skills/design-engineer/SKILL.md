---
name: design-system
description: |
  Use this skill for any frontend work: web apps, dashboards, React/Next.js projects, UI components.
  Provides color tokens, component patterns, typography, and animation rules.
  Uses shadcn/ui conventions with a monotone (black/white) aesthetic.
triggers:
  - web app
  - frontend
  - UI
  - React
  - Next.js
  - Tailwind
  - dashboard
  - components
  - shadcn
---

# Design System

A minimal, monotone design system inspired by Linear, Vercel, and Raycast.

---

## Step 1: Copy globals.css

Copy `globals.css` from this folder into your project's `src/app/globals.css`.

This file contains:

- Color tokens (shadcn-compatible)
- Typography settings with letter-spacing
- Shadow definitions
- Modal CSS classes
- Scrollbar styling

---

## Step 2: Install Dependencies

```bash
# shadcn/ui
pnpm dlx shadcn@latest init

# Icons (use Remix, never Lucide)
pnpm add react-icons

# Animations (if using card-to-modal)
pnpm add framer-motion
```

---

## Step 3: Build Components

Reference `components.md` for exact class patterns:

- Buttons (primary, secondary, ghost, destructive)
- Inputs, selects, checkboxes
- Cards (static, interactive, selected)
- Navigation items
- Sidebar and header layout
- Modal structure

---

## Step 4: Add Animations (Optional)

If building card-to-modal transitions, see `framer-motion.md` for:

- MotionConfig setup
- layoutId patterns
- Z-index management
- AnimatePresence structure

---

## Step 5: Verify

Run through `checklist.md` before completing any UI task.

---

# Rules

## ❌ Never Do This

```tsx
hover:scale-105              // No scale on hover
hover:-translate-y-1         // No translate on hover
whileHover={{ scale: 1.02 }} // No Framer Motion scale
duration-300                 // Too slow
bg-blue-500                  // No colored buttons
bg-purple-600                // No colored buttons
rounded-xl                   // Too large
focus:ring-2                 // Too thick
border-border                // Missing border width
```

## ✅ Always Do This

```tsx
hover:bg-accent              // Background change
hover:shadow-sm              // Subtle shadow
duration-100                 // Fast (100ms or 150ms)
bg-primary                   // Monotone (black/white)
text-primary-foreground      // Button text
rounded-md                   // 4px radius
focus-visible:ring-1         // Subtle ring
border border-border         // Width + color
```

---

# Quick Reference

## Colors

| Purpose         | Class                     |
| --------------- | ------------------------- |
| Page background | `bg-background`           |
| Card surface    | `bg-card`                 |
| Sidebar         | `bg-sidebar`              |
| Hover state     | `bg-accent`               |
| Primary text    | `text-foreground`         |
| Muted text      | `text-muted-foreground`   |
| Border          | `border-border`           |
| Primary button  | `bg-primary`              |
| Button text     | `text-primary-foreground` |

## Typography

| Use Case    | Size               | Tracking         |
| ----------- | ------------------ | ---------------- |
| UI text     | `text-[0.8125rem]` | `-0.006em`       |
| Small/label | `text-xs`          | `tracking-tight` |
| Card title  | `text-sm`          | `-0.006em`       |
| Section     | `text-lg`          | `-0.011em`       |
| Page title  | `text-xl`          | `-0.014em`       |
| Display     | `text-2xl`         | `-0.014em`       |

Always add `font-semibold` to titles.

## Spacing

| Token   | Value | Use              |
| ------- | ----- | ---------------- |
| `p-1.5` | 6px   | Button padding-y |
| `p-3`   | 12px  | Button padding-x |
| `p-4`   | 16px  | Card padding     |
| `p-6`   | 24px  | Section padding  |
| `gap-2` | 8px   | Tight gap        |
| `gap-4` | 16px  | Default gap      |

## Shadows

| Use Case  | Class       |
| --------- | ----------- |
| Cards     | `shadow-sm` |
| Dropdowns | `shadow-md` |
| Modals    | `shadow-lg` |

## Layout

| Element | Size/Class                      |
| ------- | ------------------------------- |
| Sidebar | `w-[244px] bg-sidebar border-r` |
| Header  | `h-10 px-6 border-b`            |
| Z-index | `z-40` (header), `z-50` (modal) |

---

# File Reference

| File               | Contains                             |
| ------------------ | ------------------------------------ |
| `globals.css`      | All CSS variables, tokens, utilities |
| `components.md`    | Copy-paste component patterns        |
| `checklist.md`     | Pre-build verification checklist     |
| `framer-motion.md` | Card-to-modal animation guide        |
