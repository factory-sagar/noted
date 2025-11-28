# Pre-Build Checklist

Run through this checklist before marking any UI task complete.

---

## Theme & Colors

- [ ] `globals.css` contains the exact theme tokens from this skill's `globals.css`
- [ ] Primary buttons use `bg-primary text-primary-foreground` (black/white, NOT colored)
- [ ] All badges use `bg-secondary text-secondary-foreground` (gray, NOT colored)
- [ ] No hardcoded hex colors in component classes

## Borders

- [ ] All borders use `border-border` or `border-input`
- [ ] Sidebar has `border-r border-sidebar-border`
- [ ] Header has `border-b border-border`
- [ ] Cards use `border border-border`

## Border Radius

- [ ] Buttons/inputs use `rounded-md`
- [ ] Pills/avatars use `rounded-full`
- [ ] No `rounded-xl` or `rounded-2xl` on controls

## Hover States

- [ ] No `hover:scale-*` transforms
- [ ] No `hover:-translate-y-*` transforms
- [ ] No `whileHover={{ scale: * }}` in Framer Motion
- [ ] Only using: `hover:bg-*`, `hover:border-*`, `hover:shadow-sm`

## Animations

- [ ] All transitions use `duration-100` or `duration-150`
- [ ] No `duration-300`, `duration-500`, or slower
- [ ] Framer Motion uses spring: `{ type: "spring", stiffness: 300, damping: 30 }`

## Layout

- [ ] Sidebar width is `w-[244px]`
- [ ] Header height is `h-10` (40px)
- [ ] Header uses `px-6` horizontal padding
- [ ] Sticky header has `z-40`

## Typography

- [ ] UI text uses `text-[0.8125rem] tracking-[-0.006em]` (13px with tight tracking)
- [ ] Titles use negative tracking: `-0.011em` for sections, `-0.014em` for pages
- [ ] Font has `font-sans antialiased` on body/root
- [ ] Primary text uses `text-foreground`
- [ ] Muted text uses `text-muted-foreground`
- [ ] Paragraphs use `text-muted-foreground` color

## Icons

- [ ] Using `react-icons/ri` (Remix Icons)
- [ ] NOT using `lucide-react`
- [ ] Default icon size is `w-4 h-4`

## Focus States (Subtle)

- [ ] Buttons use `focus-visible:ring-1 focus-visible:ring-ring`
- [ ] Inputs use `focus:border-ring` (border darkens, no ring)
- [ ] No `focus:ring-2` (too thick)
- [ ] Using `focus-visible` instead of `focus` where possible

## Modal/Overlay (if applicable)

- [ ] Using CSS classes from globals.css: `.modal-overlay`, `.modal-container`, `.modal-content`
- [ ] Overlay and modal are **SIBLINGS** in AnimatePresence (not nested)
- [ ] Card and modal share identical `layoutId`: `layoutId={\`card-${slug}\`}`
- [ ] Using `useMotionValue` for z-index (not React state)
- [ ] `borderRadius` set inline: `style={{ borderRadius: 12 }}`
- [ ] Source card hidden with placeholder div during animation
- [ ] Page wrapped with `<MotionConfig transition={{ type: "spring", stiffness: 300, damping: 30 }}>`
- [ ] No `transition-*` CSS classes on elements with `layoutId`
- [ ] Using `requestAnimationFrame` when setting z-index before selection

---

## Quick Fixes for Common Issues

| Issue               | Fix                                                  |
| ------------------- | ---------------------------------------------------- |
| Card lifts on hover | Remove `hover:-translate-y-*`, use `hover:shadow-sm` |
| Purple/blue buttons | Change to `bg-primary text-primary-foreground`       |
| Slow animations     | Change to `duration-100` or `duration-150`           |
| Missing border      | Add `border` before `border-border`                  |
| Colored badges      | Change to `bg-secondary text-secondary-foreground`   |
| Lucide icons        | Replace with `react-icons/ri` imports                |
| Large border radius | Change `rounded-xl` to `rounded-md`                  |
| Thick focus ring    | Change `focus:ring-2` to `focus-visible:ring-1`      |
| Wrong ring color    | Change to `ring-ring`                                |
| Loose text spacing  | Add `tracking-[-0.006em]` for UI, `tracking-[-0.011em]` for titles |
| Generic font        | Ensure `font-sans` is applied and Inter is loaded    |
