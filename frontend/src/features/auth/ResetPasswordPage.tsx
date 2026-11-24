import { ArrowLeft, ArrowRight, CheckCircle2, Eye, EyeOff } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { resetPassword } from './auth.api'
import { AuthPageShell, authInputClass, authPrimaryBtn } from './AuthPageShell'

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')?.trim() ?? ''

  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [done, setDone] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(null)

    if (!token) {
      setError('Liên kết đặt lại không hợp lệ. Vui lòng yêu cầu liên kết mới.')
      return
    }
    if (password.length < 8) {
      setError('Mật khẩu tối thiểu 8 ký tự.')
      return
    }
    if (password !== confirmPassword) {
      setError('Mật khẩu xác nhận không khớp.')
      return
    }

    setSubmitting(true)
    try {
      await resetPassword(token, password)
      setDone(true)
    } catch {
      setError('Không đặt lại được mật khẩu. Liên kết có thể đã hết hạn hoặc đã được dùng.')
    } finally {
      setSubmitting(false)
    }
  }

  if (done) {
    return (
      <AuthPageShell title="Đặt lại mật khẩu" subtitle="Mật khẩu đã được cập nhật.">
        <div className="flex flex-col items-center gap-4 py-4 text-center">
          <CheckCircle2 size={44} className="text-primary" />
          <p className="text-sm text-on-surface-variant">
            Bạn có thể đăng nhập bằng mật khẩu mới.
          </p>
          <Link to="/login" className={authPrimaryBtn}>
            <span>Đăng nhập</span>
            <ArrowRight size={18} />
          </Link>
        </div>
      </AuthPageShell>
    )
  }

  return (
    <AuthPageShell
      title="Đặt lại mật khẩu"
      subtitle="Chọn mật khẩu mới cho tài khoản bác sĩ của bạn."
    >
      <form className="flex flex-col space-y-3" onSubmit={handleSubmit}>
        <div className="flex flex-col space-y-1">
          <label className="px-1 text-sm font-medium text-on-surface-variant" htmlFor="new-password">
            Mật khẩu mới
          </label>
          <div className="relative">
            <input
              id="new-password"
              type={showPassword ? 'text' : 'password'}
              className={authInputClass}
              placeholder="Tối thiểu 8 ký tự"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="new-password"
              minLength={8}
              required
            />
            <button
              type="button"
              className="absolute right-4 top-1/2 flex -translate-y-1/2 items-center text-on-surface-variant"
              onClick={() => setShowPassword((v) => !v)}
              aria-label={showPassword ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'}
            >
              {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
            </button>
          </div>
        </div>

        <div className="flex flex-col space-y-1">
          <label className="px-1 text-sm font-medium text-on-surface-variant" htmlFor="confirm-password">
            Xác nhận mật khẩu
          </label>
          <input
            id="confirm-password"
            type={showPassword ? 'text' : 'password'}
            className={authInputClass}
            placeholder="Nhập lại mật khẩu mới"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            autoComplete="new-password"
            required
          />
        </div>

        {error && (
          <p className="text-sm text-error" role="alert">
            {error}
          </p>
        )}

        <button type="submit" disabled={submitting} className={`w-full ${authPrimaryBtn}`}>
          <span>{submitting ? 'Đang lưu…' : 'Đặt lại mật khẩu'}</span>
          <ArrowRight size={18} />
        </button>
      </form>

      <div className="mt-4 text-center">
        <Link
          to="/forgot-password"
          className="inline-flex items-center gap-1 text-sm text-on-surface-variant hover:underline"
        >
          <ArrowLeft size={16} />
          Yêu cầu liên kết mới
        </Link>
      </div>
    </AuthPageShell>
  )
}
