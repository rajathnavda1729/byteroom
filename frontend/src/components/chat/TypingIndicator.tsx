interface TypingUser {
  user_id: string
  display_name: string
}

interface Props {
  typingUsers: TypingUser[]
}

function typingText(users: TypingUser[]): string {
  if (users.length === 1) return `${users[0].display_name} is typing`
  if (users.length === 2) return `${users[0].display_name} and ${users[1].display_name} are typing`
  return `${users.length} people are typing`
}

export function TypingIndicator({ typingUsers }: Props) {
  if (typingUsers.length === 0) return null

  return (
    <div data-testid="typing-indicator" className="flex items-center gap-2 px-4 py-2 text-xs text-gray-400">
      <span data-testid="typing-dots" className="flex gap-0.5 items-center">
        {[0, 1, 2].map((i) => (
          <span
            key={i}
            className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce"
            style={{ animationDelay: `${i * 150}ms` }}
          />
        ))}
      </span>
      <span>{typingText(typingUsers)}</span>
    </div>
  )
}
