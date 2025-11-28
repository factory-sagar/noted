# Framer Motion Patterns

This guide covers card-to-modal transitions and other animation patterns.

---

## Required CSS (Add to globals.css)

```css
/* Modal Overlay */
.modal-overlay {
  background: rgba(0, 0, 0, 0.6);
  position: fixed;
  inset: 0;
  pointer-events: auto;
  will-change: opacity;
}

/* Modal Container - centers the modal */
.modal-container {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: auto;
  z-index: 50;
  padding: 2rem;
}

/* Modal Content - the actual modal box */
.modal-content {
  max-width: 42rem;
  width: 100%;
  max-height: 85vh;
  cursor: default;
  overflow: hidden;
  background: var(--card);
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  position: relative;
  display: flex;
  flex-direction: column;
}
```

---

## Card-to-Modal Animation

### Step 1: Page Setup

Wrap your page content with `MotionConfig` for consistent transitions:

```tsx
"use client";

import { useState } from "react";
import { motion, AnimatePresence, MotionConfig } from "framer-motion";

export default function Page() {
  const [selectedItem, setSelectedItem] = useState<Item | null>(null);
  const [initialSlug, setInitialSlug] = useState<string | null>(null);

  const handleSelect = (item: Item) => {
    setSelectedItem(item);
    setInitialSlug(item.slug);
  };

  const handleClose = () => {
    setSelectedItem(null);
    setInitialSlug(null);
  };

  return (
    <MotionConfig transition={{ type: "spring", stiffness: 300, damping: 30 }}>
      <div className="min-h-screen bg-background">
        {/* Cards Grid */}
        <div className="grid grid-cols-3 gap-4">
          {items.map((item) =>
            // Hide the card that's currently expanded
            initialSlug !== item.slug ? (
              <Card
                key={item.slug}
                item={item}
                onSelect={() => handleSelect(item)}
              />
            ) : (
              <div key={item.slug} className="w-full" /> // Placeholder
            )
          )}
        </div>

        {/* Modal - Overlay and Content as SIBLINGS */}
        <AnimatePresence>
          {selectedItem && initialSlug && (
            <>
              {/* Overlay */}
              <motion.div
                key="overlay"
                className="modal-overlay"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                onClick={handleClose}
              />
              {/* Modal */}
              <Modal
                item={selectedItem}
                initialSlug={initialSlug}
                onClose={handleClose}
              />
            </>
          )}
        </AnimatePresence>
      </div>
    </MotionConfig>
  );
}
```

### Step 2: Card Component

```tsx
"use client";

import { motion, useMotionValue } from "framer-motion";

interface CardProps {
  item: Item;
  onSelect: () => void;
}

export function Card({ item, onSelect }: CardProps) {
  // Z-index as motion value for smooth transitions
  const zIndex = useMotionValue(0);

  return (
    <motion.div
      // CRITICAL: layoutId must match the modal's layoutId
      layoutId={`card-${item.slug}`}
      onClick={() => {
        // Set z-index BEFORE triggering selection
        requestAnimationFrame(() => {
          zIndex.set(50);
          onSelect();
        });
      }}
      style={{
        zIndex,
        borderRadius: 12, // Inline for smooth animation
      }}
      // Reset z-index after animation completes
      onLayoutAnimationStart={() => zIndex.set(50)}
      onLayoutAnimationComplete={() => zIndex.set(0)}
      className="cursor-pointer border border-border bg-card overflow-hidden hover:border-ring hover:shadow-sm transition-colors duration-100"
    >
      {/* Image with its own layoutId for smooth transition */}
      <motion.div
        layoutId={`image-${item.slug}`}
        className="relative w-full h-48 overflow-hidden"
      >
        <img
          src={item.image}
          alt={item.name}
          className="absolute inset-0 w-full h-full object-cover"
        />
      </motion.div>

      <div className="p-3">
        <div className="flex items-center gap-2 mb-2">
          {/* Logo with layoutId */}
          <motion.img
            layoutId={`logo-${item.slug}`}
            src={item.logo}
            alt=""
            className="w-6 h-6 rounded-md"
          />
          {/* Name with layoutId */}
          <motion.h3
            layoutId={`name-${item.slug}`}
            className="text-sm font-semibold text-foreground"
          >
            {item.name}
          </motion.h3>
        </div>
        <p className="text-xs text-muted-foreground line-clamp-2">
          {item.description}
        </p>
      </div>
    </motion.div>
  );
}
```

### Step 3: Modal Component

```tsx
"use client";

import { motion, AnimatePresence } from "framer-motion";

interface ModalProps {
  item: Item;
  initialSlug: string;
  onClose: () => void;
}

export function Modal({ item, initialSlug, onClose }: ModalProps) {
  return (
    <div className="modal-container" onClick={onClose}>
      <motion.div
        // CRITICAL: Must match the card's layoutId
        layoutId={`card-${initialSlug}`}
        style={{ borderRadius: 12 }}
        className="modal-content"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Inner AnimatePresence for content transitions when navigating */}
        <AnimatePresence mode="wait">
          <motion.div
            key={item.slug}
            initial={{ opacity: 0, filter: "blur(4px)" }}
            animate={{ opacity: 1, filter: "blur(0px)" }}
            exit={{ opacity: 0, filter: "blur(4px)" }}
            transition={{ duration: 0.15 }}
            className="flex flex-col h-full"
          >
            {/* Cover Image - Note: NO layoutId here, different from card */}
            <div className="relative w-full h-32 overflow-hidden rounded-t-xl">
              <img
                src={item.image}
                alt=""
                className="w-full h-full object-cover"
              />
            </div>

            {/* Close Button */}
            <button
              onClick={onClose}
              className="absolute top-3 right-3 w-7 h-7 flex items-center justify-center bg-background/90 hover:bg-background rounded-full"
            >
              <RiCloseLine className="w-4 h-4" />
            </button>

            {/* Content */}
            <div className="p-4 overflow-y-auto">
              <div className="flex items-start gap-3 mb-4">
                <img src={item.logo} alt="" className="w-10 h-10 rounded-lg" />
                <div>
                  <h2 className="text-lg font-bold text-foreground">
                    {item.name}
                  </h2>
                  <p className="text-xs text-muted-foreground">
                    {item.description}
                  </p>
                </div>
              </div>

              {/* Rest of modal content */}
              <div className="space-y-4">{/* ... */}</div>
            </div>
          </motion.div>
        </AnimatePresence>
      </motion.div>
    </div>
  );
}
```

---

## Critical Rules

### 1. layoutId Matching

The card and modal MUST share the same `layoutId`:

```tsx
// Card
<motion.div layoutId={`card-${item.slug}`}>

// Modal
<motion.div layoutId={`card-${initialSlug}`}>
```

### 2. Z-Index Management

Use `useMotionValue` for z-index to avoid React re-renders:

```tsx
const zIndex = useMotionValue(0);

// Set BEFORE triggering modal
requestAnimationFrame(() => {
  zIndex.set(50);
  onSelect();
});
```

### 3. AnimatePresence Structure

Overlay and modal must be **siblings** inside AnimatePresence:

```tsx
<AnimatePresence>
  {selected && (
    <>
      <motion.div key="overlay" ... />
      <Modal ... />
    </>
  )}
</AnimatePresence>
```

**NEVER nest them:**

```tsx
// ❌ WRONG
<AnimatePresence>
  {selected && (
    <motion.div className="overlay">
      <Modal /> {/* Nested = broken animation */}
    </motion.div>
  )}
</AnimatePresence>
```

### 4. No CSS Transitions on layoutId Elements

Never add Tailwind `transition-*` classes to elements with `layoutId`:

```tsx
// ❌ WRONG - CSS transition conflicts with Framer Motion
<motion.div layoutId="card" className="transition-all duration-200">

// ✅ CORRECT
<motion.div layoutId="card" className="border border-border">
```

### 5. Inline borderRadius

Always set `borderRadius` inline for smooth animation:

```tsx
<motion.div
  layoutId="card"
  style={{ borderRadius: 12 }}
>
```

### 6. MotionConfig for Consistent Springs

Wrap your page with MotionConfig:

```tsx
<MotionConfig transition={{ type: "spring", stiffness: 300, damping: 30 }}>
  {/* Your content */}
</MotionConfig>
```

### 7. Hide Source Card During Animation

Replace the source card with a placeholder while modal is open:

```tsx
{
  items.map((item) =>
    initialSlug !== item.slug ? (
      <Card key={item.slug} item={item} />
    ) : (
      <div key={item.slug} className="w-full" /> // Invisible placeholder
    )
  );
}
```

---

## Escape Key Handler

```tsx
useEffect(() => {
  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape" && selectedItem) {
      handleClose();
    }
  }
  window.addEventListener("keydown", handleKeyDown);
  return () => window.removeEventListener("keydown", handleKeyDown);
}, [selectedItem]);
```

---

## Navigation Between Items (Optional)

If you want arrow key navigation in the modal:

```tsx
const handleNext = () => {
  const currentIndex = items.findIndex((i) => i.slug === selectedItem?.slug);
  const nextIndex = (currentIndex + 1) % items.length;
  setSelectedItem(items[nextIndex]);
  // Note: Don't change initialSlug - keeps the layoutId animation working
};

const handlePrevious = () => {
  const currentIndex = items.findIndex((i) => i.slug === selectedItem?.slug);
  const prevIndex = currentIndex === 0 ? items.length - 1 : currentIndex - 1;
  setSelectedItem(items[prevIndex]);
};
```

---

## Common Issues

| Issue                           | Cause                                  | Fix                                               |
| ------------------------------- | -------------------------------------- | ------------------------------------------------- |
| Modal doesn't animate from card | `layoutId` mismatch                    | Ensure both use identical `layoutId`              |
| Animation is janky              | CSS `transition-*` on layoutId element | Remove all CSS transitions from layoutId elements |
| Z-index flicker                 | React re-render                        | Use `useMotionValue` for z-index                  |
| Modal appears instantly         | Missing `AnimatePresence`              | Wrap with `AnimatePresence`                       |
| Content jumps during animation  | Missing inline `borderRadius`          | Add `style={{ borderRadius: 12 }}`                |
| Card visible behind modal       | Not hiding source card                 | Replace card with placeholder div                 |
