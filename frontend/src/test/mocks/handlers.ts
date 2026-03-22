import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/users/me', () => {
    return HttpResponse.json({
      user_id: 'user-1',
      username: 'alice',
      display_name: 'Alice',
    })
  }),

  http.get('/api/chats', () => {
    return HttpResponse.json({ chats: [] })
  }),
]
