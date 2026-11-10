import { roleLabels } from '../auth/role-labels'
import { useAuth } from '../auth/use-auth'
import type { User } from '../lib/auth-api'

export function AccountView({ user }: { user: User }) {
  return (
    <section aria-labelledby="account-title" className="mx-auto w-full max-w-md">
      <h1 id="account-title" className="text-3xl font-bold tracking-tight text-slate-950">
        Tài khoản
      </h1>
      <dl className="mt-6 border border-slate-300 bg-white p-5 text-sm">
        <dt className="font-semibold text-slate-800">Email</dt>
        <dd className="mt-1 break-all text-slate-700">{user.email}</dd>
        <dt className="mt-4 font-semibold text-slate-800">Vai trò</dt>
        <dd className="mt-1 text-slate-700">{roleLabels[user.role]}</dd>
      </dl>
    </section>
  )
}

export function AccountPage() {
  const { state } = useAuth()
  if (state.status !== 'authenticated') return null
  return <AccountView user={state.user} />
}
