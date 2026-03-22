import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { api } from '@/services/api'

interface RegisterForm {
  username: string
  password: string
  confirmPassword: string
}

interface AuthUserDTO {
  user_id: string
  username: string
  display_name: string
  avatar_url?: string
}

interface RegisterResponse {
  token: string
  expires_at: string
  user: AuthUserDTO
}

export function RegisterPage() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<RegisterForm>()

  const password = watch('password')

  const onSubmit = async (data: RegisterForm) => {
    setServerError(null)
    try {
      const res = await api.post<RegisterResponse>('/api/auth/register', {
        username: data.username,
        password: data.password,
      })
      setAuth({
        user: {
          id: res.user.user_id,
          username: res.user.username,
          displayName: res.user.display_name,
          avatarUrl: res.user.avatar_url ?? null,
        },
        token: res.token,
      })
      navigate('/')
    } catch (err: unknown) {
      setServerError(err instanceof Error ? err.message : 'Registration failed')
    }
  }

  return (
    <main className="min-h-screen flex items-center justify-center bg-gray-950 px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-white">ByteRoom</h1>
          <p className="text-gray-400 mt-2">Create your account</p>
        </div>

        <form
          onSubmit={handleSubmit(onSubmit)}
          className="bg-gray-900 rounded-2xl p-8 shadow-2xl space-y-5"
          noValidate
        >
          {serverError && (
            <div role="alert" className="bg-red-900/40 border border-red-700 text-red-300 rounded-lg px-4 py-3 text-sm">
              {serverError}
            </div>
          )}

          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-300 mb-1.5">
              Username
            </label>
            <input
              id="username"
              type="text"
              autoComplete="username"
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-2.5 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition"
              placeholder="alice"
              {...register('username', {
                required: 'Username is required',
                minLength: { value: 3, message: 'At least 3 characters' },
                pattern: { value: /^[a-z0-9_]+$/, message: 'Lowercase letters, numbers, underscores only' },
              })}
            />
            {errors.username && (
              <p role="alert" className="text-red-400 text-xs mt-1.5">
                {errors.username.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-300 mb-1.5">
              Password
            </label>
            <input
              id="password"
              type="password"
              autoComplete="new-password"
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-2.5 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition"
              placeholder="••••••••"
              {...register('password', {
                required: 'Password is required',
                minLength: { value: 8, message: 'At least 8 characters' },
              })}
            />
            {errors.password && (
              <p role="alert" className="text-red-400 text-xs mt-1.5">
                {errors.password.message}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-300 mb-1.5">
              Confirm Password
            </label>
            <input
              id="confirmPassword"
              type="password"
              autoComplete="new-password"
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-2.5 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition"
              placeholder="••••••••"
              {...register('confirmPassword', {
                required: 'Please confirm your password',
                validate: (v) => v === password || 'Passwords must match',
              })}
            />
            {errors.confirmPassword && (
              <p role="alert" className="text-red-400 text-xs mt-1.5">
                {errors.confirmPassword.message}
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full bg-blue-600 hover:bg-blue-500 disabled:bg-blue-800 disabled:cursor-not-allowed text-white font-semibold rounded-lg py-2.5 transition"
          >
            {isSubmitting ? 'Creating account…' : 'Sign Up'}
          </button>

          <p className="text-center text-sm text-gray-400">
            Already have an account?{' '}
            <Link to="/login" className="text-blue-400 hover:text-blue-300">
              Sign in
            </Link>
          </p>
        </form>
      </div>
    </main>
  )
}
