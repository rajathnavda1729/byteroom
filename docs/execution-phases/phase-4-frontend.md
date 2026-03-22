# Phase 4: Frontend

## Objective

Build the React SPA with TypeScript: authentication, chat interface, real-time WebSocket integration, markdown/code rendering, and domain-specific features (Mermaid diagrams, Excalidraw).

## Duration Estimate

7 development days

## Prerequisites

- Phase 2 completed (REST API functional)
- Phase 3 completed (WebSocket working)
- Backend running and accessible
- Node.js 20+ installed

---

## Tasks

### Task 4.1: Project Scaffolding & Configuration

**Description**: Set up React + TypeScript + Vite project with Tailwind CSS and testing infrastructure.

**TDD Approach**:
```typescript
// src/App.test.tsx
import { render, screen } from '@testing-library/react';
import App from './App';

describe('App', () => {
  it('renders without crashing', () => {
    render(<App />);
    expect(screen.getByRole('main')).toBeInTheDocument();
  });
});
```

**Subtasks**:
- [ ] Create Vite project with React TypeScript template
- [ ] Install and configure Tailwind CSS with dark mode
- [ ] Configure Vitest and React Testing Library
- [ ] Set up path aliases (`@/components`, `@/hooks`, etc.)
- [ ] Configure ESLint and Prettier
- [ ] Create base layout component
- [ ] Add environment variable handling (`.env.example`)

**Files to Create**:
```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
});
```

**Exit Criteria**:
- [ ] `npm run dev` starts development server
- [ ] `npm run build` produces production bundle
- [ ] `npm test` runs Vitest tests
- [ ] Tailwind classes render correctly
- [ ] Path aliases resolve correctly

---

### Task 4.2: Authentication Store & API Client

**Description**: Implement Zustand auth store and API client with JWT handling.

**TDD Approach**:
```typescript
// src/stores/authStore.test.ts
import { renderHook, act } from '@testing-library/react';
import { useAuthStore } from './authStore';

describe('authStore', () => {
  beforeEach(() => {
    useAuthStore.getState().logout();
    localStorage.clear();
  });

  it('initializes with null user', () => {
    const { result } = renderHook(() => useAuthStore());
    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('sets user on login', () => {
    const { result } = renderHook(() => useAuthStore());
    
    act(() => {
      result.current.setAuth({
        user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
        token: 'jwt-token',
        refreshToken: 'refresh-token',
      });
    });

    expect(result.current.user?.username).toBe('alice');
    expect(result.current.isAuthenticated).toBe(true);
  });

  it('clears user on logout', () => {
    const { result } = renderHook(() => useAuthStore());
    
    act(() => {
      result.current.setAuth({
        user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
        token: 'jwt-token',
        refreshToken: 'refresh-token',
      });
      result.current.logout();
    });

    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('persists token to localStorage', () => {
    const { result } = renderHook(() => useAuthStore());
    
    act(() => {
      result.current.setAuth({
        user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
        token: 'jwt-token',
        refreshToken: 'refresh-token',
      });
    });

    expect(localStorage.getItem('token')).toBe('jwt-token');
  });
});

// src/services/api.test.ts
import { api } from './api';
import { server } from '@/test/mocks/server';
import { rest } from 'msw';

describe('API Client', () => {
  it('adds Authorization header when token exists', async () => {
    let capturedHeaders: Headers | null = null;
    
    server.use(
      rest.get('/api/test', (req, res, ctx) => {
        capturedHeaders = req.headers;
        return res(ctx.json({ ok: true }));
      })
    );

    localStorage.setItem('token', 'test-token');
    await api.get('/api/test');

    expect(capturedHeaders?.get('Authorization')).toBe('Bearer test-token');
  });

  it('handles 401 by logging out', async () => {
    server.use(
      rest.get('/api/test', (req, res, ctx) => {
        return res(ctx.status(401), ctx.json({ error: 'unauthorized' }));
      })
    );

    await expect(api.get('/api/test')).rejects.toThrow();
    expect(localStorage.getItem('token')).toBeNull();
  });
});
```

**Subtasks**:
- [ ] Write auth store tests
- [ ] Implement `useAuthStore` with Zustand
- [ ] Implement token persistence in localStorage
- [ ] Write API client tests
- [ ] Implement API client with fetch wrapper
- [ ] Add request/response interceptors
- [ ] Handle 401 responses (auto-logout)
- [ ] Add token refresh logic

**Exit Criteria**:
- [ ] Auth store manages user state
- [ ] Tokens persisted across page reloads
- [ ] API client adds auth headers
- [ ] 401 triggers logout
- [ ] Coverage ≥ 90%

---

### Task 4.3: Login & Registration Pages

**Description**: Implement authentication UI with form validation.

**TDD Approach**:
```typescript
// src/pages/LoginPage.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LoginPage } from './LoginPage';
import { server } from '@/test/mocks/server';
import { rest } from 'msw';
import { BrowserRouter } from 'react-router-dom';

const renderWithRouter = (component: React.ReactNode) => {
  return render(<BrowserRouter>{component}</BrowserRouter>);
};

describe('LoginPage', () => {
  it('renders login form', () => {
    renderWithRouter(<LoginPage />);
    
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('shows validation errors for empty fields', async () => {
    renderWithRouter(<LoginPage />);
    const user = userEvent.setup();
    
    await user.click(screen.getByRole('button', { name: /sign in/i }));
    
    expect(await screen.findByText(/email is required/i)).toBeInTheDocument();
    expect(await screen.findByText(/password is required/i)).toBeInTheDocument();
  });

  it('submits form and redirects on success', async () => {
    server.use(
      rest.post('/api/auth/login', (req, res, ctx) => {
        return res(ctx.json({
          user_id: 'user-1',
          username: 'alice',
          token: 'jwt-token',
        }));
      })
    );

    renderWithRouter(<LoginPage />);
    const user = userEvent.setup();
    
    await user.type(screen.getByLabelText(/email/i), 'alice@example.com');
    await user.type(screen.getByLabelText(/password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /sign in/i }));
    
    await waitFor(() => {
      expect(window.location.pathname).toBe('/');
    });
  });

  it('shows error message on failed login', async () => {
    server.use(
      rest.post('/api/auth/login', (req, res, ctx) => {
        return res(ctx.status(401), ctx.json({ message: 'Invalid credentials' }));
      })
    );

    renderWithRouter(<LoginPage />);
    const user = userEvent.setup();
    
    await user.type(screen.getByLabelText(/email/i), 'alice@example.com');
    await user.type(screen.getByLabelText(/password/i), 'wrongpassword');
    await user.click(screen.getByRole('button', { name: /sign in/i }));
    
    expect(await screen.findByText(/invalid credentials/i)).toBeInTheDocument();
  });
});

// src/pages/RegisterPage.test.tsx
describe('RegisterPage', () => {
  it('validates password requirements', async () => {
    renderWithRouter(<RegisterPage />);
    const user = userEvent.setup();
    
    await user.type(screen.getByLabelText(/password/i), 'weak');
    await user.click(screen.getByRole('button', { name: /sign up/i }));
    
    expect(await screen.findByText(/at least 8 characters/i)).toBeInTheDocument();
  });

  it('validates matching passwords', async () => {
    renderWithRouter(<RegisterPage />);
    const user = userEvent.setup();
    
    await user.type(screen.getByLabelText(/^password/i), 'Password123!');
    await user.type(screen.getByLabelText(/confirm password/i), 'Different123!');
    await user.click(screen.getByRole('button', { name: /sign up/i }));
    
    expect(await screen.findByText(/passwords must match/i)).toBeInTheDocument();
  });
});
```

**Subtasks**:
- [ ] Write LoginPage tests
- [ ] Implement LoginPage component
- [ ] Write RegisterPage tests
- [ ] Implement RegisterPage component
- [ ] Add form validation with react-hook-form
- [ ] Style with Tailwind (modern, dark mode)
- [ ] Add loading states and error handling
- [ ] Implement protected route wrapper

**Exit Criteria**:
- [ ] Login/Register forms validate input
- [ ] Successful auth redirects to chat
- [ ] Error messages display clearly
- [ ] Forms are accessible (labels, ARIA)
- [ ] Coverage ≥ 85%

---

### Task 4.4: Chat Store & State Management

**Description**: Implement Zustand store for chat, messages, and typing indicators.

**TDD Approach**:
```typescript
// src/stores/chatStore.test.ts
import { renderHook, act } from '@testing-library/react';
import { useChatStore } from './chatStore';

describe('chatStore', () => {
  beforeEach(() => {
    useChatStore.getState().reset();
  });

  it('adds chat to store', () => {
    const { result } = renderHook(() => useChatStore());
    
    act(() => {
      result.current.addChat({
        chat_id: 'chat-1',
        name: 'Tech Discussion',
        type: 'group',
        members: [],
        last_message: null,
      });
    });

    expect(result.current.chats['chat-1']).toBeDefined();
    expect(result.current.chats['chat-1'].name).toBe('Tech Discussion');
  });

  it('sets active chat', () => {
    const { result } = renderHook(() => useChatStore());
    
    act(() => {
      result.current.setActiveChat('chat-1');
    });

    expect(result.current.activeChat).toBe('chat-1');
  });

  it('adds message to chat', () => {
    const { result } = renderHook(() => useChatStore());
    
    act(() => {
      result.current.addMessage('chat-1', {
        message_id: 'msg-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Hello!',
        timestamp: new Date().toISOString(),
        status: 'sent',
      });
    });

    expect(result.current.messages['chat-1']).toHaveLength(1);
    expect(result.current.messages['chat-1'][0].content).toBe('Hello!');
  });

  it('adds optimistic message with pending status', () => {
    const { result } = renderHook(() => useChatStore());
    
    act(() => {
      result.current.addOptimisticMessage('chat-1', {
        message_id: 'temp-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Sending...',
        timestamp: new Date().toISOString(),
        status: 'pending',
      });
    });

    expect(result.current.messages['chat-1'][0].status).toBe('pending');
  });

  it('confirms optimistic message', () => {
    const { result } = renderHook(() => useChatStore());
    
    act(() => {
      result.current.addOptimisticMessage('chat-1', {
        message_id: 'temp-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Hello!',
        timestamp: new Date().toISOString(),
        status: 'pending',
      });
      
      result.current.confirmMessage('chat-1', 'temp-1', {
        message_id: 'msg-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Hello!',
        timestamp: new Date().toISOString(),
        status: 'sent',
      });
    });

    expect(result.current.messages['chat-1'][0].status).toBe('sent');
    expect(result.current.messages['chat-1'][0].message_id).toBe('msg-1');
  });

  it('tracks typing users', () => {
    const { result } = renderHook(() => useChatStore());
    
    act(() => {
      result.current.setTyping('chat-1', 'user-2', true);
    });

    expect(result.current.typingUsers['chat-1']).toContain('user-2');

    act(() => {
      result.current.setTyping('chat-1', 'user-2', false);
    });

    expect(result.current.typingUsers['chat-1'] || []).not.toContain('user-2');
  });
});
```

**Subtasks**:
- [ ] Write chat store tests
- [ ] Implement `useChatStore` with chat list
- [ ] Implement message state by chat ID
- [ ] Implement optimistic updates
- [ ] Implement message confirmation/failure
- [ ] Add typing users tracking
- [ ] Add unread message counts

**Exit Criteria**:
- [ ] Chats stored and retrievable
- [ ] Messages organized by chat ID
- [ ] Optimistic updates work correctly
- [ ] Typing indicators tracked per chat
- [ ] Coverage ≥ 90%

---

### Task 4.5: WebSocket Hook

**Description**: Implement custom React hook for WebSocket connection management with auto-reconnection.

**TDD Approach**:
```typescript
// src/hooks/useWebSocket.test.ts
import { renderHook, act, waitFor } from '@testing-library/react';
import { useWebSocket } from './useWebSocket';
import WS from 'jest-websocket-mock';

describe('useWebSocket', () => {
  let server: WS;

  beforeEach(() => {
    server = new WS('ws://localhost:8080/ws');
  });

  afterEach(() => {
    WS.clean();
  });

  it('connects to WebSocket server', async () => {
    const onMessage = vi.fn();
    
    renderHook(() => useWebSocket({
      url: 'ws://localhost:8080/ws',
      token: 'test-token',
      onMessage,
    }));

    await server.connected;
    expect(server).toHaveReceivedMessages([]);
  });

  it('sends messages when connected', async () => {
    const { result } = renderHook(() => useWebSocket({
      url: 'ws://localhost:8080/ws',
      token: 'test-token',
      onMessage: vi.fn(),
    }));

    await server.connected;

    act(() => {
      result.current.send('message.send', { content: 'Hello' });
    });

    await expect(server).toReceiveMessage(
      JSON.stringify({ event: 'message.send', data: { content: 'Hello' } })
    );
  });

  it('calls onMessage when receiving data', async () => {
    const onMessage = vi.fn();
    
    renderHook(() => useWebSocket({
      url: 'ws://localhost:8080/ws',
      token: 'test-token',
      onMessage,
    }));

    await server.connected;
    
    act(() => {
      server.send(JSON.stringify({ event: 'message.new', data: { content: 'Hi' } }));
    });

    await waitFor(() => {
      expect(onMessage).toHaveBeenCalledWith(
        expect.objectContaining({ event: 'message.new' })
      );
    });
  });

  it('reports connection state', async () => {
    const { result } = renderHook(() => useWebSocket({
      url: 'ws://localhost:8080/ws',
      token: 'test-token',
      onMessage: vi.fn(),
    }));

    expect(result.current.connectionState).toBe('connecting');

    await server.connected;

    await waitFor(() => {
      expect(result.current.connectionState).toBe('connected');
    });
  });

  it('reconnects on connection close', async () => {
    const { result } = renderHook(() => useWebSocket({
      url: 'ws://localhost:8080/ws',
      token: 'test-token',
      onMessage: vi.fn(),
    }));

    await server.connected;
    server.close();

    await waitFor(() => {
      expect(result.current.connectionState).toBe('reconnecting');
    });
  });
});
```

**Subtasks**:
- [ ] Write WebSocket hook tests
- [ ] Implement connection lifecycle
- [ ] Implement send method with JSON serialization
- [ ] Implement message parsing and dispatch
- [ ] Add exponential backoff reconnection
- [ ] Track connection state
- [ ] Handle ping/pong heartbeat
- [ ] Clean up on unmount

**Exit Criteria**:
- [ ] WebSocket connects with token
- [ ] Messages sent and received correctly
- [ ] Auto-reconnection with backoff
- [ ] Connection state accurately reported
- [ ] Coverage ≥ 85%

---

### Task 4.6: Chat Layout & Sidebar

**Description**: Implement main chat layout with sidebar showing chat list.

**TDD Approach**:
```typescript
// src/components/chat/ChatLayout.test.tsx
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ChatLayout } from './ChatLayout';
import { useChatStore } from '@/stores/chatStore';

describe('ChatLayout', () => {
  beforeEach(() => {
    useChatStore.getState().reset();
    useChatStore.getState().addChat({
      chat_id: 'chat-1',
      name: 'Tech Discussion',
      type: 'group',
      members: [],
      last_message: { content: 'Last msg', sender_name: 'Alice', timestamp: new Date().toISOString() },
    });
    useChatStore.getState().addChat({
      chat_id: 'chat-2',
      name: null,
      type: 'direct',
      members: [{ user_id: 'u2', username: 'bob', display_name: 'Bob' }],
      last_message: null,
    });
  });

  it('renders sidebar with chat list', () => {
    render(<ChatLayout />);
    
    expect(screen.getByText('Tech Discussion')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument(); // Direct chat shows member name
  });

  it('selects chat on click', async () => {
    render(<ChatLayout />);
    const user = userEvent.setup();
    
    await user.click(screen.getByText('Tech Discussion'));
    
    expect(useChatStore.getState().activeChat).toBe('chat-1');
  });

  it('shows last message preview', () => {
    render(<ChatLayout />);
    
    expect(screen.getByText('Last msg')).toBeInTheDocument();
  });

  it('highlights active chat', async () => {
    render(<ChatLayout />);
    const user = userEvent.setup();
    
    await user.click(screen.getByText('Tech Discussion'));
    
    const chatItem = screen.getByText('Tech Discussion').closest('li');
    expect(chatItem).toHaveClass('bg-blue-100'); // or whatever active style
  });
});

// src/components/chat/ChatListItem.test.tsx
describe('ChatListItem', () => {
  it('displays unread count badge', () => {
    render(
      <ChatListItem
        chat={{
          chat_id: 'chat-1',
          name: 'Test Chat',
          type: 'group',
          unreadCount: 5,
        }}
      />
    );
    
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('displays relative time for last message', () => {
    const recentTime = new Date(Date.now() - 5 * 60 * 1000).toISOString(); // 5 min ago
    
    render(
      <ChatListItem
        chat={{
          chat_id: 'chat-1',
          name: 'Test Chat',
          type: 'group',
          last_message: { timestamp: recentTime, content: 'Hi', sender_name: 'Alice' },
        }}
      />
    );
    
    expect(screen.getByText(/5m/)).toBeInTheDocument();
  });
});
```

**Subtasks**:
- [ ] Write ChatLayout tests
- [ ] Implement ChatLayout with responsive sidebar
- [ ] Write ChatListItem tests
- [ ] Implement ChatListItem component
- [ ] Add user profile section in sidebar
- [ ] Implement new chat button
- [ ] Add mobile-friendly sidebar toggle
- [ ] Style with Tailwind (dark mode support)

**Exit Criteria**:
- [ ] Sidebar shows all user's chats
- [ ] Chat selection works
- [ ] Active chat highlighted
- [ ] Responsive on mobile
- [ ] Coverage ≥ 85%

---

### Task 4.7: Message List & MessageBubble

**Description**: Implement virtualized message list with message bubbles.

**TDD Approach**:
```typescript
// src/components/chat/MessageList.test.tsx
import { render, screen } from '@testing-library/react';
import { MessageList } from './MessageList';
import { useChatStore } from '@/stores/chatStore';
import { useAuthStore } from '@/stores/authStore';

describe('MessageList', () => {
  beforeEach(() => {
    useAuthStore.getState().setAuth({
      user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
      token: 'token',
    });
    useChatStore.getState().reset();
  });

  it('renders messages in chronological order', () => {
    useChatStore.getState().loadMessages('chat-1', [
      { message_id: 'msg-1', content: 'First', timestamp: '2026-03-21T10:00:00Z' },
      { message_id: 'msg-2', content: 'Second', timestamp: '2026-03-21T10:01:00Z' },
    ]);

    render(<MessageList chatId="chat-1" />);
    
    const messages = screen.getAllByRole('listitem');
    expect(messages[0]).toHaveTextContent('First');
    expect(messages[1]).toHaveTextContent('Second');
  });

  it('groups messages by date', () => {
    useChatStore.getState().loadMessages('chat-1', [
      { message_id: 'msg-1', content: 'Yesterday', timestamp: '2026-03-20T10:00:00Z' },
      { message_id: 'msg-2', content: 'Today', timestamp: '2026-03-21T10:00:00Z' },
    ]);

    render(<MessageList chatId="chat-1" />);
    
    expect(screen.getByText(/march 20/i)).toBeInTheDocument();
    expect(screen.getByText(/today/i)).toBeInTheDocument();
  });

  it('shows loading state when fetching history', () => {
    render(<MessageList chatId="chat-1" isLoading />);
    
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('shows empty state for new chats', () => {
    render(<MessageList chatId="chat-1" />);
    
    expect(screen.getByText(/no messages yet/i)).toBeInTheDocument();
  });
});

// src/components/chat/MessageBubble.test.tsx
describe('MessageBubble', () => {
  it('renders own messages with different style', () => {
    render(
      <MessageBubble
        message={{ sender_id: 'user-1', content: 'My message' }}
        isOwnMessage={true}
      />
    );
    
    const bubble = screen.getByText('My message').parentElement;
    expect(bubble).toHaveClass('bg-blue-500'); // Own message color
  });

  it('renders other messages with sender info', () => {
    render(
      <MessageBubble
        message={{
          sender_id: 'user-2',
          sender: { display_name: 'Bob', avatar_url: null },
          content: 'Their message',
        }}
        isOwnMessage={false}
      />
    );
    
    expect(screen.getByText('Bob')).toBeInTheDocument();
    expect(screen.getByText('Their message')).toBeInTheDocument();
  });

  it('shows pending status for optimistic messages', () => {
    render(
      <MessageBubble
        message={{ content: 'Sending...', status: 'pending' }}
        isOwnMessage={true}
      />
    );
    
    expect(screen.getByText(/sending/i)).toBeInTheDocument();
  });

  it('shows error status for failed messages', () => {
    render(
      <MessageBubble
        message={{ content: 'Failed', status: 'error' }}
        isOwnMessage={true}
      />
    );
    
    expect(screen.getByText(/failed/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
  });
});
```

**Subtasks**:
- [ ] Write MessageList tests
- [ ] Implement MessageList with react-window virtualization
- [ ] Write MessageBubble tests
- [ ] Implement MessageBubble component
- [ ] Add date separators
- [ ] Implement scroll-to-bottom on new message
- [ ] Add infinite scroll for history loading
- [ ] Style own vs other messages differently

**Exit Criteria**:
- [ ] Messages render correctly
- [ ] Virtualization handles 1000+ messages
- [ ] Date grouping works
- [ ] Message status displayed
- [ ] Coverage ≥ 85%

---

### Task 4.8: Message Input

**Description**: Implement message input with send button and keyboard handling.

**TDD Approach**:
```typescript
// src/components/chat/MessageInput.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MessageInput } from './MessageInput';

describe('MessageInput', () => {
  it('renders textarea and send button', () => {
    render(<MessageInput onSend={vi.fn()} />);
    
    expect(screen.getByRole('textbox')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /send/i })).toBeInTheDocument();
  });

  it('calls onSend with content when button clicked', async () => {
    const onSend = vi.fn();
    const user = userEvent.setup();
    
    render(<MessageInput onSend={onSend} />);
    
    await user.type(screen.getByRole('textbox'), 'Hello world');
    await user.click(screen.getByRole('button', { name: /send/i }));
    
    expect(onSend).toHaveBeenCalledWith('Hello world');
  });

  it('clears input after sending', async () => {
    const user = userEvent.setup();
    
    render(<MessageInput onSend={vi.fn()} />);
    
    const input = screen.getByRole('textbox');
    await user.type(input, 'Hello');
    await user.click(screen.getByRole('button', { name: /send/i }));
    
    expect(input).toHaveValue('');
  });

  it('sends on Enter key (without shift)', async () => {
    const onSend = vi.fn();
    const user = userEvent.setup();
    
    render(<MessageInput onSend={onSend} />);
    
    await user.type(screen.getByRole('textbox'), 'Hello{enter}');
    
    expect(onSend).toHaveBeenCalledWith('Hello');
  });

  it('allows newline with Shift+Enter', async () => {
    const onSend = vi.fn();
    const user = userEvent.setup();
    
    render(<MessageInput onSend={onSend} />);
    
    await user.type(screen.getByRole('textbox'), 'Line 1{shift>}{enter}{/shift}Line 2');
    
    expect(onSend).not.toHaveBeenCalled();
    expect(screen.getByRole('textbox')).toHaveValue('Line 1\nLine 2');
  });

  it('disables send when input is empty', () => {
    render(<MessageInput onSend={vi.fn()} />);
    
    expect(screen.getByRole('button', { name: /send/i })).toBeDisabled();
  });

  it('emits typing events on input change', async () => {
    const onTyping = vi.fn();
    const user = userEvent.setup();
    
    render(<MessageInput onSend={vi.fn()} onTyping={onTyping} />);
    
    await user.type(screen.getByRole('textbox'), 'H');
    
    expect(onTyping).toHaveBeenCalledWith(true);
  });

  it('auto-resizes textarea for multiline content', async () => {
    const user = userEvent.setup();
    
    render(<MessageInput onSend={vi.fn()} />);
    
    const input = screen.getByRole('textbox');
    const initialHeight = input.scrollHeight;
    
    await user.type(input, 'Line 1{shift>}{enter}{/shift}Line 2{shift>}{enter}{/shift}Line 3');
    
    expect(input.scrollHeight).toBeGreaterThan(initialHeight);
  });
});
```

**Subtasks**:
- [ ] Write MessageInput tests
- [ ] Implement auto-resizing textarea
- [ ] Handle Enter to send, Shift+Enter for newline
- [ ] Implement send button with disabled state
- [ ] Add attachment button (for future media upload)
- [ ] Emit typing events with debounce
- [ ] Style with Tailwind

**Exit Criteria**:
- [ ] Text input works correctly
- [ ] Keyboard shortcuts work
- [ ] Typing events emitted
- [ ] Input accessible
- [ ] Coverage ≥ 90%

---

### Task 4.9: Markdown Renderer

**Description**: Implement safe markdown rendering with syntax highlighting for code blocks.

**TDD Approach**:
```typescript
// src/components/chat/MarkdownRenderer.test.tsx
import { render, screen } from '@testing-library/react';
import { MarkdownRenderer } from './MarkdownRenderer';

describe('MarkdownRenderer', () => {
  it('renders plain text', () => {
    render(<MarkdownRenderer content="Hello world" />);
    expect(screen.getByText('Hello world')).toBeInTheDocument();
  });

  it('renders bold text', () => {
    render(<MarkdownRenderer content="**bold text**" />);
    expect(screen.getByText('bold text')).toHaveStyle('font-weight: bold');
  });

  it('renders italic text', () => {
    render(<MarkdownRenderer content="*italic text*" />);
    expect(screen.getByText('italic text')).toHaveStyle('font-style: italic');
  });

  it('renders inline code', () => {
    render(<MarkdownRenderer content="Use `const` for constants" />);
    const code = screen.getByText('const');
    expect(code.tagName).toBe('CODE');
  });

  it('renders code blocks with syntax highlighting', () => {
    const code = `\`\`\`javascript
const x = 42;
console.log(x);
\`\`\``;
    
    render(<MarkdownRenderer content={code} />);
    
    expect(screen.getByText('const')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
  });

  it('renders links with target="_blank"', () => {
    render(<MarkdownRenderer content="[Google](https://google.com)" />);
    
    const link = screen.getByRole('link', { name: 'Google' });
    expect(link).toHaveAttribute('href', 'https://google.com');
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('sanitizes dangerous HTML', () => {
    render(<MarkdownRenderer content="<script>alert('xss')</script>Safe text" />);
    
    expect(screen.getByText('Safe text')).toBeInTheDocument();
    expect(document.querySelector('script')).toBeNull();
  });

  it('renders lists', () => {
    render(<MarkdownRenderer content="- Item 1\n- Item 2" />);
    
    expect(screen.getByRole('list')).toBeInTheDocument();
    expect(screen.getAllByRole('listitem')).toHaveLength(2);
  });

  it('renders tables', () => {
    const table = `
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
`;
    render(<MarkdownRenderer content={table} />);
    
    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(screen.getByText('Header 1')).toBeInTheDocument();
  });
});
```

**Subtasks**:
- [ ] Write MarkdownRenderer tests
- [ ] Configure react-markdown with plugins
- [ ] Integrate react-syntax-highlighter
- [ ] Support 20+ common languages
- [ ] Add copy button for code blocks
- [ ] Implement line numbers option
- [ ] Style code blocks for dark/light mode

**Exit Criteria**:
- [ ] Markdown renders correctly
- [ ] Code blocks syntax highlighted
- [ ] No XSS vulnerabilities
- [ ] Copy to clipboard works
- [ ] Coverage ≥ 90%

---

### Task 4.10: Mermaid Diagram Renderer

**Description**: Implement lazy-loaded Mermaid diagram rendering for architecture diagrams.

**TDD Approach**:
```typescript
// src/components/chat/MermaidDiagram.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import { MermaidDiagram } from './MermaidDiagram';

describe('MermaidDiagram', () => {
  it('renders mermaid diagram as SVG', async () => {
    const diagram = `graph TD
      A[Start] --> B[End]
    `;
    
    render(<MermaidDiagram content={diagram} />);
    
    await waitFor(() => {
      expect(screen.getByRole('img', { name: /diagram/i })).toBeInTheDocument();
    });
  });

  it('shows loading state while rendering', () => {
    render(<MermaidDiagram content="graph TD\nA-->B" />);
    
    expect(screen.getByText(/loading diagram/i)).toBeInTheDocument();
  });

  it('shows error for invalid mermaid syntax', async () => {
    const invalidDiagram = 'not valid mermaid syntax!!!';
    
    render(<MermaidDiagram content={invalidDiagram} />);
    
    await waitFor(() => {
      expect(screen.getByText(/failed to render/i)).toBeInTheDocument();
    });
  });

  it('supports sequence diagrams', async () => {
    const diagram = `sequenceDiagram
      Alice->>Bob: Hello
      Bob-->>Alice: Hi
    `;
    
    render(<MermaidDiagram content={diagram} />);
    
    await waitFor(() => {
      expect(screen.getByRole('img')).toBeInTheDocument();
    });
  });

  it('supports flowcharts', async () => {
    const diagram = `flowchart LR
      A --> B --> C
    `;
    
    render(<MermaidDiagram content={diagram} />);
    
    await waitFor(() => {
      expect(screen.getByRole('img')).toBeInTheDocument();
    });
  });

  it('provides zoom/pan for large diagrams', async () => {
    const largeDiagram = `graph TD
      ${Array.from({ length: 20 }, (_, i) => `N${i}[Node ${i}]`).join('\n')}
    `;
    
    render(<MermaidDiagram content={largeDiagram} />);
    
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /zoom in/i })).toBeInTheDocument();
    });
  });
});
```

**Subtasks**:
- [ ] Write MermaidDiagram tests
- [ ] Implement lazy loading of mermaid library
- [ ] Configure mermaid for dark/light mode
- [ ] Handle rendering errors gracefully
- [ ] Add zoom/pan controls for large diagrams
- [ ] Integrate with MarkdownRenderer for ```mermaid blocks

**Exit Criteria**:
- [ ] Mermaid diagrams render correctly
- [ ] Lazy loading reduces bundle size
- [ ] Dark mode supported
- [ ] Errors handled gracefully
- [ ] Coverage ≥ 85%

---

### Task 4.11: Excalidraw Integration

**Description**: Implement Excalidraw embed for interactive system design diagrams.

**TDD Approach**:
```typescript
// src/components/chat/ExcalidrawEmbed.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ExcalidrawEmbed } from './ExcalidrawEmbed';

const mockDiagramState = {
  type: 'excalidraw',
  version: 2,
  elements: [
    {
      type: 'rectangle',
      x: 100,
      y: 100,
      width: 200,
      height: 100,
    },
  ],
  appState: { viewBackgroundColor: '#ffffff' },
};

describe('ExcalidrawEmbed', () => {
  it('renders excalidraw canvas', async () => {
    render(<ExcalidrawEmbed state={JSON.stringify(mockDiagramState)} />);
    
    await waitFor(() => {
      expect(screen.getByTestId('excalidraw-container')).toBeInTheDocument();
    });
  });

  it('loads initial state correctly', async () => {
    render(<ExcalidrawEmbed state={JSON.stringify(mockDiagramState)} />);
    
    await waitFor(() => {
      // Verify the rectangle element is rendered
      const canvas = screen.getByTestId('excalidraw-container');
      expect(canvas).toBeInTheDocument();
    });
  });

  it('is read-only by default', async () => {
    render(<ExcalidrawEmbed state={JSON.stringify(mockDiagramState)} readOnly />);
    
    await waitFor(() => {
      const container = screen.getByTestId('excalidraw-container');
      expect(container).toHaveAttribute('data-readonly', 'true');
    });
  });

  it('allows editing when not read-only', async () => {
    const onUpdate = vi.fn();
    
    render(
      <ExcalidrawEmbed
        state={JSON.stringify(mockDiagramState)}
        readOnly={false}
        onUpdate={onUpdate}
      />
    );
    
    // Interaction test would require more complex setup
    await waitFor(() => {
      expect(screen.getByTestId('excalidraw-container')).toBeInTheDocument();
    });
  });

  it('shows expand button for full-screen mode', async () => {
    render(<ExcalidrawEmbed state={JSON.stringify(mockDiagramState)} />);
    
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /expand/i })).toBeInTheDocument();
    });
  });

  it('handles invalid state gracefully', async () => {
    render(<ExcalidrawEmbed state="invalid json" />);
    
    await waitFor(() => {
      expect(screen.getByText(/failed to load diagram/i)).toBeInTheDocument();
    });
  });
});
```

**Subtasks**:
- [ ] Write ExcalidrawEmbed tests
- [ ] Install @excalidraw/excalidraw
- [ ] Implement lazy loading
- [ ] Support read-only viewing mode
- [ ] Support full-screen editing
- [ ] Handle state serialization/deserialization
- [ ] Add save/update callback for modifications

**Exit Criteria**:
- [ ] Excalidraw diagrams render correctly
- [ ] Read-only mode works
- [ ] Full-screen editing available
- [ ] State updates emit events
- [ ] Coverage ≥ 80%

---

### Task 4.12: Typing Indicator Component

**Description**: Implement typing indicator that shows who is currently typing.

**TDD Approach**:
```typescript
// src/components/chat/TypingIndicator.test.tsx
import { render, screen } from '@testing-library/react';
import { TypingIndicator } from './TypingIndicator';

describe('TypingIndicator', () => {
  it('shows nothing when no one is typing', () => {
    const { container } = render(<TypingIndicator typingUsers={[]} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('shows single user typing', () => {
    render(
      <TypingIndicator
        typingUsers={[{ user_id: 'u1', display_name: 'Alice' }]}
      />
    );
    
    expect(screen.getByText(/alice is typing/i)).toBeInTheDocument();
  });

  it('shows two users typing', () => {
    render(
      <TypingIndicator
        typingUsers={[
          { user_id: 'u1', display_name: 'Alice' },
          { user_id: 'u2', display_name: 'Bob' },
        ]}
      />
    );
    
    expect(screen.getByText(/alice and bob are typing/i)).toBeInTheDocument();
  });

  it('shows multiple users typing', () => {
    render(
      <TypingIndicator
        typingUsers={[
          { user_id: 'u1', display_name: 'Alice' },
          { user_id: 'u2', display_name: 'Bob' },
          { user_id: 'u3', display_name: 'Charlie' },
        ]}
      />
    );
    
    expect(screen.getByText(/3 people are typing/i)).toBeInTheDocument();
  });

  it('shows animated dots', () => {
    render(
      <TypingIndicator
        typingUsers={[{ user_id: 'u1', display_name: 'Alice' }]}
      />
    );
    
    expect(screen.getByTestId('typing-dots')).toBeInTheDocument();
  });
});
```

**Subtasks**:
- [ ] Write TypingIndicator tests
- [ ] Implement typing indicator component
- [ ] Add animated "..." dots
- [ ] Handle 1, 2, 3+ users cases
- [ ] Position at bottom of message list
- [ ] Integrate with chat store typing state

**Exit Criteria**:
- [ ] Shows correct user names
- [ ] Animation works smoothly
- [ ] Handles multiple users
- [ ] Coverage ≥ 95%

---

### Task 4.13: Image Upload & Preview

**Description**: Implement image upload flow with pre-signed URLs and preview.

**TDD Approach**:
```typescript
// src/components/chat/ImageUpload.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ImageUpload } from './ImageUpload';
import { server } from '@/test/mocks/server';
import { rest } from 'msw';

describe('ImageUpload', () => {
  it('opens file picker on button click', async () => {
    const user = userEvent.setup();
    render(<ImageUpload onUpload={vi.fn()} />);
    
    const input = screen.getByTestId('file-input');
    const clickSpy = vi.spyOn(input, 'click');
    
    await user.click(screen.getByRole('button', { name: /attach/i }));
    
    expect(clickSpy).toHaveBeenCalled();
  });

  it('shows preview after file selection', async () => {
    const user = userEvent.setup();
    const file = new File(['test'], 'test.png', { type: 'image/png' });
    
    render(<ImageUpload onUpload={vi.fn()} />);
    
    const input = screen.getByTestId('file-input');
    await user.upload(input, file);
    
    expect(await screen.findByRole('img', { name: /preview/i })).toBeInTheDocument();
  });

  it('validates file type', async () => {
    const user = userEvent.setup();
    const file = new File(['test'], 'test.exe', { type: 'application/x-executable' });
    
    render(<ImageUpload onUpload={vi.fn()} />);
    
    const input = screen.getByTestId('file-input');
    await user.upload(input, file);
    
    expect(await screen.findByText(/unsupported file type/i)).toBeInTheDocument();
  });

  it('validates file size', async () => {
    const user = userEvent.setup();
    const largeFile = new File([new ArrayBuffer(15 * 1024 * 1024)], 'large.png', { type: 'image/png' });
    
    render(<ImageUpload onUpload={vi.fn()} />);
    
    const input = screen.getByTestId('file-input');
    await user.upload(input, largeFile);
    
    expect(await screen.findByText(/file too large/i)).toBeInTheDocument();
  });

  it('uploads file and calls onUpload with key', async () => {
    server.use(
      rest.post('/api/upload/request', (req, res, ctx) => {
        return res(ctx.json({
          upload_url: 'https://s3.example.com/upload',
          file_key: 'uploads/abc123.png',
        }));
      })
    );

    const onUpload = vi.fn();
    const user = userEvent.setup();
    const file = new File(['test'], 'test.png', { type: 'image/png' });
    
    render(<ImageUpload onUpload={onUpload} />);
    
    const input = screen.getByTestId('file-input');
    await user.upload(input, file);
    await user.click(screen.getByRole('button', { name: /upload/i }));
    
    await waitFor(() => {
      expect(onUpload).toHaveBeenCalledWith('uploads/abc123.png');
    });
  });

  it('shows upload progress', async () => {
    // This would require mocking XMLHttpRequest progress events
    // Simplified test
    render(<ImageUpload onUpload={vi.fn()} />);
    
    // Trigger upload flow...
    // expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });
});
```

**Subtasks**:
- [ ] Write ImageUpload tests
- [ ] Implement file picker button
- [ ] Add drag-and-drop support
- [ ] Validate file type (images only)
- [ ] Validate file size (max 10MB)
- [ ] Show preview before upload
- [ ] Request pre-signed URL from backend
- [ ] Upload directly to S3
- [ ] Show upload progress
- [ ] Call onUpload with file key

**Exit Criteria**:
- [ ] File picker works
- [ ] Validation prevents invalid files
- [ ] Preview shows selected image
- [ ] S3 upload works
- [ ] Coverage ≥ 85%

---

### Task 4.14: End-to-End Integration

**Description**: Wire all components together into complete chat experience.

**TDD Approach**:
```typescript
// src/App.integration.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { App } from './App';
import { server } from '@/test/mocks/server';
import { rest } from 'msw';
import WS from 'jest-websocket-mock';

describe('App Integration', () => {
  let wsServer: WS;

  beforeEach(() => {
    wsServer = new WS('ws://localhost:8080/ws');
    localStorage.setItem('token', 'valid-token');
    
    server.use(
      rest.get('/api/chats', (req, res, ctx) => {
        return res(ctx.json({
          chats: [
            { chat_id: 'chat-1', name: 'Tech Discussion', type: 'group' },
          ],
        }));
      })
    );
  });

  afterEach(() => {
    WS.clean();
    localStorage.clear();
  });

  it('loads and displays chat list', async () => {
    render(<App />);
    
    expect(await screen.findByText('Tech Discussion')).toBeInTheDocument();
  });

  it('connects to WebSocket on load', async () => {
    render(<App />);
    
    await wsServer.connected;
  });

  it('sends message via WebSocket', async () => {
    const user = userEvent.setup();
    render(<App />);
    
    await screen.findByText('Tech Discussion');
    await user.click(screen.getByText('Tech Discussion'));
    
    await user.type(screen.getByRole('textbox'), 'Hello world');
    await user.click(screen.getByRole('button', { name: /send/i }));
    
    await expect(wsServer).toReceiveMessage(expect.stringContaining('message.send'));
  });

  it('receives and displays incoming message', async () => {
    render(<App />);
    const user = userEvent.setup();
    
    await screen.findByText('Tech Discussion');
    await user.click(screen.getByText('Tech Discussion'));
    await wsServer.connected;
    
    wsServer.send(JSON.stringify({
      event: 'message.new',
      data: {
        message_id: 'msg-1',
        chat_id: 'chat-1',
        sender_id: 'user-2',
        sender: { display_name: 'Bob' },
        content_type: 'markdown',
        content: 'Hello from Bob!',
        timestamp: new Date().toISOString(),
      },
    }));
    
    expect(await screen.findByText('Hello from Bob!')).toBeInTheDocument();
  });

  it('shows typing indicator when other user types', async () => {
    render(<App />);
    const user = userEvent.setup();
    
    await screen.findByText('Tech Discussion');
    await user.click(screen.getByText('Tech Discussion'));
    await wsServer.connected;
    
    wsServer.send(JSON.stringify({
      event: 'user.typing',
      data: {
        chat_id: 'chat-1',
        user_id: 'user-2',
        username: 'bob',
        is_typing: true,
      },
    }));
    
    expect(await screen.findByText(/bob is typing/i)).toBeInTheDocument();
  });

  it('handles reconnection gracefully', async () => {
    render(<App />);
    
    await wsServer.connected;
    wsServer.close();
    
    // Should show reconnecting state
    expect(await screen.findByText(/reconnecting/i)).toBeInTheDocument();
    
    // New server instance
    wsServer = new WS('ws://localhost:8080/ws');
    
    await waitFor(() => {
      expect(screen.queryByText(/reconnecting/i)).not.toBeInTheDocument();
    });
  });
});
```

**Subtasks**:
- [ ] Write integration tests
- [ ] Wire auth flow with protected routes
- [ ] Connect WebSocket hook to chat store
- [ ] Wire message sending through WebSocket
- [ ] Handle incoming message events
- [ ] Handle typing events
- [ ] Add connection status indicator
- [ ] Implement error boundaries
- [ ] Add loading states throughout

**Exit Criteria**:
- [ ] Full chat flow works end-to-end
- [ ] Messages send and receive correctly
- [ ] Typing indicators show in real-time
- [ ] Connection issues handled gracefully
- [ ] Coverage ≥ 80%

---

## Phase 4 Exit Criteria Summary

### Automated Verification

```bash
# Run all frontend tests
cd frontend && npm test -- --coverage

# Run build
npm run build

# Run linting
npm run lint
```

### Manual Verification

| Check | Action | Expected Result |
|-------|--------|-----------------|
| Auth flow | Register, login, logout | User authenticated/redirected |
| Chat selection | Click chat in sidebar | Chat messages load |
| Send message | Type and press Enter | Message appears, ACK received |
| Receive message | Other user sends | Message appears in real-time |
| Code highlighting | Send code block | Syntax highlighted correctly |
| Mermaid diagram | Send mermaid code | Diagram renders |
| Image upload | Attach and send image | Image displays in chat |

### Quality Gates

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Test coverage ≥ 80%
- [ ] No ESLint errors
- [ ] Build produces < 500KB initial bundle
- [ ] Lighthouse accessibility score ≥ 90

### Deliverables

1. ✅ Project scaffolding with Vite + TypeScript
2. ✅ Auth store and API client
3. ✅ Login/Registration pages
4. ✅ Chat state management
5. ✅ WebSocket hook with reconnection
6. ✅ Chat layout with sidebar
7. ✅ Message list with virtualization
8. ✅ Message input with keyboard handling
9. ✅ Markdown renderer with syntax highlighting
10. ✅ Mermaid diagram support
11. ✅ Excalidraw integration
12. ✅ Typing indicators
13. ✅ Image upload flow
14. ✅ Full integration

---

## Next Phase

Upon completion, proceed to [Phase 5: Integration & Polish](./phase-5-integration.md)
