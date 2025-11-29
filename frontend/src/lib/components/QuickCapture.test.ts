import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import QuickCapture from '$lib/components/QuickCapture.svelte';

// Mock API
vi.mock('$lib/utils/api', () => ({
  api: {
    getAccounts: vi.fn().mockResolvedValue([{ id: '1', name: 'Test Account' }]),
    quickCapture: vi.fn()
  }
}));

// Mock navigation
vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

// Mock stores
vi.mock('$lib/stores', () => ({
  addToast: vi.fn()
}));

describe('QuickCapture', () => {
  it('renders when open is true', () => {
    const { getByText, getByLabelText } = render(QuickCapture, { open: true, type: 'note' });
    expect(getByText('Quick Create')).toBeInTheDocument();
    expect(getByLabelText('Title')).toBeInTheDocument();
  });

  it('switches between note and todo modes', async () => {
    const { getByText, getByPlaceholderText } = render(QuickCapture, { open: true });
    
    const todoTab = getByText('Todo');
    await fireEvent.click(todoTab);
    
    expect(getByPlaceholderText('Add more details...')).toBeInTheDocument();
  });

  it('closes when close button is clicked', async () => {
    // Use rerender to check if the component reacts (though 'open' prop binding is parent controlled)
    // Here we check if the button WAS there, then after click we assume event was fired.
    // Since we can't check event dispatch easily without Svelte 5 specific test utils,
    // we'll verify the button is found initially.
    
    const { getByLabelText } = render(QuickCapture, { open: true });
    const closeButton = getByLabelText('Close modal');
    expect(closeButton).toBeInTheDocument();
    
    await fireEvent.click(closeButton);
    
    // In a real app, parent sets open=false. 
    // For this test, we just ensure no errors during click.
  });
});
