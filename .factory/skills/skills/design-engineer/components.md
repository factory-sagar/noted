# Component Patterns

Copy these exact class combinations. Do not modify.

---

## Typography Classes

Use these tracking values for proper letter-spacing:

| Use Case    | Classes                                      |
| ----------- | -------------------------------------------- |
| UI text     | `text-[0.8125rem] tracking-[-0.006em]`       |
| Small/label | `text-xs tracking-tight`                     |
| Card title  | `text-sm font-semibold tracking-[-0.006em]`  |
| Section     | `text-lg font-semibold tracking-[-0.011em]`  |
| Page title  | `text-xl font-semibold tracking-[-0.014em]`  |
| Display     | `text-2xl font-semibold tracking-[-0.014em]` |

---

## Buttons

### Primary Button (Black/White)

```tsx
<button className="px-3 py-1.5 bg-primary text-primary-foreground rounded-md text-[0.8125rem] tracking-[-0.006em] font-medium hover:bg-primary/90 transition-colors duration-100 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring">
  Button Text
</button>
```

### Secondary Button

```tsx
<button className="px-3 py-1.5 border border-border bg-background text-foreground rounded-md text-[0.8125rem] tracking-[-0.006em] font-medium hover:bg-accent transition-colors duration-100 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring">
  Secondary
</button>
```

### Ghost Button

```tsx
<button className="px-3 py-1.5 text-muted-foreground rounded-md text-[0.8125rem] tracking-[-0.006em] font-medium hover:bg-accent hover:text-foreground transition-colors duration-100 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring">
  Ghost
</button>
```

### Icon Button

```tsx
<button className="p-1.5 text-muted-foreground rounded-md hover:bg-accent hover:text-foreground transition-colors duration-100 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring">
  <RiSearchLine className="w-4 h-4" />
</button>
```

### Destructive Button

```tsx
<button className="px-3 py-1.5 bg-destructive text-destructive-foreground rounded-md text-[0.8125rem] tracking-[-0.006em] font-medium hover:bg-destructive/90 transition-opacity duration-100 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-destructive/50">
  Delete
</button>
```

---

## Inputs

### Text Input

```tsx
<input
  type="text"
  className="w-full px-3 py-1.5 bg-background border border-input rounded-md text-foreground text-[0.8125rem] tracking-[-0.006em] placeholder:text-muted-foreground focus:outline-none focus:border-ring transition-colors duration-100"
  placeholder="Enter text..."
/>
```

### Textarea

```tsx
<textarea
  className="w-full px-3 py-1.5 bg-background border border-input rounded-md text-foreground text-[0.8125rem] tracking-[-0.006em] placeholder:text-muted-foreground focus:outline-none focus:border-ring transition-colors duration-100 resize-none"
  rows={4}
  placeholder="Enter description..."
/>
```

### Select

```tsx
<select className="w-full px-3 py-1.5 bg-background border border-input rounded-md text-foreground text-[0.8125rem] tracking-[-0.006em] focus:outline-none focus:border-ring transition-colors duration-100">
  <option>Select option...</option>
</select>
```

### Checkbox

```tsx
<input
  type="checkbox"
  className="w-3.5 h-3.5 rounded border border-input text-primary focus:ring-1 focus:ring-offset-0 focus:ring-ring"
/>
```

### Form Label

```tsx
<label className="block text-xs tracking-tight font-medium text-muted-foreground mb-1.5">
  Label Text
</label>
```

---

## Cards

### Static Card

```tsx
<div className="bg-card border border-border rounded-md p-4 shadow-sm">
  <h3 className="text-sm font-semibold tracking-[-0.006em] text-card-foreground mb-2">
    Title
  </h3>
  <p className="text-[0.8125rem] tracking-[-0.006em] text-muted-foreground leading-relaxed">
    Content
  </p>
</div>
```

### Interactive Card

```tsx
<div className="bg-card border border-border rounded-md p-4 hover:border-ring hover:shadow-sm transition-all duration-100 cursor-pointer">
  <h3 className="text-sm font-semibold tracking-[-0.006em] text-card-foreground mb-2">
    Title
  </h3>
  <p className="text-[0.8125rem] tracking-[-0.006em] text-muted-foreground leading-relaxed">
    Content
  </p>
</div>
```

### Elevated Card

```tsx
<div className="bg-card border border-border rounded-md p-4 shadow-md">
  <h3 className="text-sm font-semibold text-card-foreground mb-2">Title</h3>
  <p className="text-sm text-muted-foreground leading-relaxed">Content</p>
</div>
```

### Selected Card

```tsx
<div className="bg-accent border border-ring rounded-md p-4">
  <h3 className="text-sm font-semibold text-accent-foreground mb-2">Title</h3>
  <p className="text-sm text-muted-foreground leading-relaxed">Content</p>
</div>
```

---

## Badge

### Monotone Badge (ONLY use this)

```tsx
<span className="px-2 py-0.5 bg-secondary text-secondary-foreground rounded-md text-xs">
  Badge
</span>
```

### Pill Badge

```tsx
<span className="px-2 py-0.5 bg-secondary text-secondary-foreground rounded-full text-xs">
  Pill
</span>
```

---

## Navigation

### Nav Item (Default)

```tsx
<button className="w-full px-3 py-1.5 text-left text-[0.8125rem] tracking-[-0.006em] font-medium text-muted-foreground rounded-md hover:bg-accent hover:text-foreground transition-colors duration-100">
  Nav Item
</button>
```

### Nav Item (Active)

```tsx
<button className="w-full px-3 py-1.5 text-left text-[0.8125rem] tracking-[-0.006em] font-medium text-foreground bg-accent rounded-md">
  Active Item
</button>
```

---

## Layout

### Page Container

```tsx
<div className="font-sans antialiased bg-background text-foreground min-h-screen">
  {/* content */}
</div>
```

### Sidebar (244px)

```tsx
<aside className="w-[244px] bg-sidebar border-r border-sidebar-border flex flex-col">
  {/* Sidebar header */}
  <div className="h-10 px-6 border-b border-sidebar-border flex items-center">
    <span className="text-[0.8125rem] font-semibold tracking-[-0.006em] text-sidebar-foreground">
      Title
    </span>
  </div>
  {/* Sidebar content */}
  <nav className="flex-1 p-2 overflow-y-auto">{/* Nav items */}</nav>
</aside>
```

### Header (40px)

```tsx
<header className="h-10 px-6 border-b border-border bg-background sticky top-0 z-40 flex items-center justify-between">
  <h1 className="text-[0.8125rem] font-semibold tracking-[-0.006em]">
    Page Title
  </h1>
  <div className="flex gap-2">{/* Actions */}</div>
</header>
```

### Two-Column Layout

```tsx
<div className="flex h-screen">
  {/* Sidebar */}
  <aside className="w-[244px] bg-sidebar border-r border-sidebar-border flex flex-col">
    {/* ... */}
  </aside>
  {/* Main */}
  <main className="flex-1 overflow-auto flex flex-col">
    <header className="h-10 px-6 border-b border-border bg-background sticky top-0 z-40 flex items-center">
      {/* ... */}
    </header>
    <div className="flex-1 p-6 overflow-y-auto">{/* Content */}</div>
  </main>
</div>
```

---

## Modal

### Modal Overlay

```tsx
<motion.div
  className="fixed inset-0 bg-black/60 z-50"
  initial={{ opacity: 0 }}
  animate={{ opacity: 1 }}
  exit={{ opacity: 0 }}
  transition={{ duration: 0.15 }}
  onClick={onClose}
/>
```

### Modal Container

```tsx
<div className="fixed inset-0 z-50 flex items-center justify-center p-4">
  <div className="w-full max-w-lg bg-card border border-border rounded-lg shadow-lg">
    {/* Modal header */}
    <div className="h-10 px-6 border-b border-border flex items-center justify-between">
      <h2 className="text-[0.8125rem] font-semibold tracking-[-0.006em]">
        Modal Title
      </h2>
      <button className="p-1.5 text-muted-foreground rounded-md hover:bg-accent hover:text-foreground transition-colors duration-100">
        <RiCloseLine className="w-4 h-4" />
      </button>
    </div>
    {/* Modal body */}
    <div className="p-6">{/* Content */}</div>
    {/* Modal footer */}
    <div className="px-6 py-4 border-t border-border flex justify-end gap-2">
      <button className="px-3 py-1.5 text-muted-foreground rounded-md text-[0.8125rem] tracking-[-0.006em] font-medium hover:bg-accent transition-colors duration-100">
        Cancel
      </button>
      <button className="px-3 py-1.5 bg-primary text-primary-foreground rounded-md text-[0.8125rem] tracking-[-0.006em] font-medium hover:bg-primary/90 transition-colors duration-100">
        Confirm
      </button>
    </div>
  </div>
</div>
```

---

## Grid Patterns

### Responsive Card Grid

```tsx
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  {/* Cards */}
</div>
```

### Auto-fit Grid

```tsx
<div className="grid grid-cols-[repeat(auto-fit,minmax(280px,1fr))] gap-4">
  {/* Cards */}
</div>
```

### Stats Grid

```tsx
<div className="grid grid-cols-2 lg:grid-cols-4 gap-4">{/* Stat cards */}</div>
```

---

## Dividers

### Horizontal Divider

```tsx
<div className="h-px bg-border my-4" />
```

### Vertical Divider

```tsx
<div className="w-px bg-border mx-2 h-4" />
```
