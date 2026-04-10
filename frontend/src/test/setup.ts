import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach, afterAll, beforeAll, vi } from 'vitest';
import { server } from './msw/server';

// Vitest sets import.meta.env.DEV=true regardless of the `define` config.
// Stub with empty string (falsy) so formatErrorMessage behaves like production.
vi.stubEnv('DEV', '');

// MSW matches absolute URLs used in handlers; keep in sync with tests.
vi.stubEnv('VITE_API_BASE_URL', 'http://localhost:8080');

// MSW server lifecycle
beforeAll(() => server.listen({ onUnhandledRequest: 'warn' }));
afterEach(() => {
  cleanup();
  server.resetHandlers();
});
afterAll(() => server.close());

// Polyfill: window.matchMedia (MUI uses it)
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Polyfill: ResizeObserver (MUI DataGrid / responsive components)
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
window.ResizeObserver = ResizeObserverStub as unknown as typeof ResizeObserver;

// Polyfill: URL.createObjectURL / revokeObjectURL (report export)
if (typeof URL.createObjectURL === 'undefined') {
  URL.createObjectURL = vi.fn(() => 'blob:http://localhost/fake-url');
}
if (typeof URL.revokeObjectURL === 'undefined') {
  URL.revokeObjectURL = vi.fn();
}

// Polyfill: AudioContext (SecurityPage feedback sounds)
class AudioContextStub {
  createOscillator() {
    return {
      connect: vi.fn(),
      frequency: { value: 0 },
      type: 'sine',
      start: vi.fn(),
      stop: vi.fn(),
    };
  }
  createGain() {
    return {
      connect: vi.fn(),
      gain: {
        value: 0,
        setValueAtTime: vi.fn(),
        exponentialRampToValueAtTime: vi.fn(),
      },
    };
  }
  get destination() {
    return {};
  }
  get currentTime() {
    return 0;
  }
}
(window as any).AudioContext = AudioContextStub;
(window as any).webkitAudioContext = AudioContextStub;

// Polyfill: scrollTo
window.scrollTo = vi.fn() as any;

// Polyfill: HTMLElement.scrollIntoView
Element.prototype.scrollIntoView = vi.fn();
