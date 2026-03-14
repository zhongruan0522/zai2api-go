'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { toast } from 'sonner';

export default function LoginPage() {
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !password) return;
    setLoading(true);

    try {
      await login(username, password);
      toast.success('登录成功');
      router.push('/tokens');
    } catch (err) {
      toast.error('登录失败', {
        description: err instanceof Error ? err.message : '用户名或密码错误',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen px-4">
      <div className="flex min-h-screen items-center justify-center">
        <div className="w-full max-w-md rounded-[2.5rem] border border-border bg-card p-10 shadow-2xl shadow-black/10">
          <div className="text-center">
            <h1 className="text-3xl font-semibold text-foreground">ZAI2API</h1>
            <p className="mt-2 text-sm text-muted-foreground">管理员登录</p>
          </div>

          <form onSubmit={handleSubmit} className="mt-6 space-y-4">
            <div className="space-y-2">
              <label className="block text-sm font-medium text-foreground">用户名</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="admin"
                disabled={loading}
                required
                className="w-full rounded-2xl border border-input bg-background px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div className="space-y-2">
              <label className="block text-sm font-medium text-foreground">密码</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                disabled={loading}
                required
                className="w-full rounded-2xl border border-input bg-background px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            <button
              type="submit"
              disabled={loading || !username || !password}
              className="w-full rounded-2xl bg-primary py-3 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {loading ? '登录中...' : '登录'}
            </button>
          </form>

          <div className="mt-6 rounded-2xl bg-secondary/60 p-3">
            <p className="text-xs text-muted-foreground">
              ZAI2API · 仅限授权人员使用
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
