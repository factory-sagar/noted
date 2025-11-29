import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import CommandPalette from '$lib/components/CommandPalette.svelte';

// Mock navigation
vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

describe('CommandPalette', () => {
  it('renders when open is true', () => {
    const { getByPlaceholderText } = render(CommandPalette, { open: true });
    expect(getByPlaceholderText('Type a command or search...')).toBeInTheDocument();
  });

  it('does not render when open is false', () => {
    const { queryByPlaceholderText } = render(CommandPalette, { open: false });
    expect(queryByPlaceholderText('Type a command or search...')).not.toBeInTheDocument();
  });

  it('shows static commands by default', () => {
    const { getByText } = render(CommandPalette, { open: true });
    expect(getByText('Go to Notes')).toBeInTheDocument();
    expect(getByText('New Note')).toBeInTheDocument();
  });

  it('filters commands based on input', async () => {
    const { getByPlaceholderText, getByText, queryByText } = render(CommandPalette, { open: true });
    const input = getByPlaceholderText('Type a command or search...');
    
    await fireEvent.input(input, { target: { value: 'Settings' } });
    
    expect(getByText('Go to Settings')).toBeInTheDocument();
    expect(queryByText('Go to Notes')).not.toBeInTheDocument();
  });
});
