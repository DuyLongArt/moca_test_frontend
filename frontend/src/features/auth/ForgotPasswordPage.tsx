import { ArrowLeft, ArrowRight, MailCheck } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { requestPasswordReset } from './auth.api'
import { AuthPageShell, authInputClass, authPrimaryBtn } from './AuthPageShell'

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [submitted, setSubmitted] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await requestPasswordReset(email.trim())
      setSubmitted(true)
    } catch {
      setError('Không gửi được liên kết đặt lại. Vui lòng thử lại sau.')
    } finally {
      setSubmitting(false)
    }
  }

  if (submitted) {
    return (
      <AuthPageShell
        title="Kiểm tra email"
        subtitle="Chúng tôi đã gửi liên kết đặt lại mật khẩu."
      >
        <div className="flex flex-col items-center gap-4 py-4 text-center">
          <MailCheck size={44} className="text-primary" />
          <p className="text-sm text-on-surface-variant">
            Nếu tài khoản tồn tại, bạn sẽ nhận được liên kết đặt lại mật khẩu trong email.
            Liên kết có hiệu lực trong 30 phút.
          </p>
          <Link to="/login" className={authPrimaryBtn}>
            <span>Về đăng nhập</span>
            <ArrowRight size={18} />
          </Link>
        </div>
      </AuthPageShell>
    )
  }

  return (
    <AuthPageShell
      title="Quên mật khẩu"
      subtitle="Nhập email bác sĩ để nhận liên kết đặt lại mật khẩu."
    >
      <form className="flex flex-col space-y-3" onSubmit={handleSubmit}>
        <div className="flex flex-col space-y-1">
          <label className="px-1 text-sm font-medium text-on-surface-variant" htmlFor="reset-email">
            Email
          </label>
          <input
            id="reset-email"
            type="email"
            className={authInputClass}
            placeholder="doctor@benhvien.vn"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
            required
          />
        </div>

        {error && (
          <p className="text-sm text-error" role="alert">
            {error}
          </p>
        )}

        <button type="submit" disabled={submitting} className={`w-full ${authPrimaryBtn}`}>
          <span>{submitting ? 'Đang gửi…' : 'Gửi liên kết đặt lại'}</span>
          <ArrowRight size={18} />
        </button>
      </form>

      <div className="mt-4 text-center">
        <Link
          to="/login"
          className="inline-flex items-center gap-1 text-sm text-on-surface-variant hover:underline"
        >
          <ArrowLeft size={16} />
          Về đăng nhập
        </Link>
      </div>
    </AuthPageShell>
  )
}
