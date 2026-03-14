'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import {
  LayoutDashboard,
  Users,
  KeyRound,
  FileText,
  Activity,
  LogOut,
  PanelLeftClose,
  PanelLeft,
  RefreshCw,
} from 'lucide-react';

const menuItems = [
  {
    path: '/tokens',
    label: 'Token 管理',
    icon: Users,
  },
  {
    path: '/apikeys',
    label: 'APIKey 管理',
    icon: KeyRound,
  },
  {
    path: '/logs',
    label: '运行日志',
    icon: FileText,
  },
  {
    path: '/monitor',
    label: '服务监控',
    icon: Activity,
  },
];

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { isAuthenticated, logout } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  useEffect(() => {
    const saved = localStorage.getItem('sidebar-collapsed');
    if (saved) setCollapsed(saved === 'true');
  }, []);

  useEffect(() => {
    localStorage.setItem('sidebar-collapsed', String(collapsed));
  }, [collapsed]);

  // Close mobile sidebar on route change
  useEffect(() => {
    setSidebarOpen(false);
  }, [pathname]);

  const currentPageTitle = menuItems.find((item) => item.path === pathname)?.label || 'Token 管理';

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen">
      <div className="flex min-h-screen flex-col lg:flex-row">
        {/* Mobile overlay */}
        {sidebarOpen && (
          <div
            className="fixed inset-0 z-30 bg-black/20 backdrop-blur-sm lg:hidden"
            onClick={() => setSidebarOpen(false)}
          />
        )}

        {/* Sidebar */}
        <aside
          className={`fixed inset-y-0 left-0 z-40 border-r border-border bg-card/90 backdrop-blur-sm transition-[width,transform] duration-200 ease-out will-change-[transform] transform-gpu flex flex-col lg:static lg:translate-x-0 lg:bg-card/80 lg:border-b-0 lg:sticky lg:top-0 lg:h-screen ${
            collapsed ? 'w-20' : 'w-72'
          } -translate-x-full lg:translate-x-0`}
          style={sidebarOpen ? { transform: 'translateX(0)' } : undefined}
        >
          {/* Logo */}
          <div className={`flex h-16 items-center justify-between px-6 pt-4 lg:h-20 lg:pt-5 ${collapsed ? 'justify-center px-0' : ''}`}>
            <div className={`flex items-center gap-2 ${collapsed ? 'gap-0 justify-center w-full' : ''}`}>
              <span className={`text-base font-semibold text-foreground ${collapsed ? 'text-xs' : ''}`}>
                {collapsed ? 'ZAI' : 'ZAI2API'}
              </span>
            </div>
          </div>

          {/* Navigation */}
          <nav className={`flex-1 overflow-y-auto pb-4 pt-4 lg:pt-6 ${collapsed ? 'px-2' : 'px-3'}`}>
            {!collapsed && (
              <p className="px-3 pb-2 text-xs uppercase tracking-[0.28em] text-muted-foreground">
                导航
              </p>
            )}
            <div className="space-y-1">
              {menuItems.map((item) => {
                const isActive = pathname === item.path;
                const Icon = item.icon;
                return (
                  <Link
                    key={item.path}
                    href={item.path}
                    className={`group flex items-center rounded-2xl py-2 text-sm font-medium transition-colors overflow-hidden ${
                      isActive
                        ? 'bg-accent text-foreground'
                        : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                    } ${collapsed ? 'px-2 justify-center gap-0' : 'px-3 gap-3'}`}
                    title={collapsed ? item.label : undefined}
                  >
                    <span
                      className={`inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-2xl border border-border ${
                        isActive
                          ? 'bg-primary/15 text-primary border-primary/30'
                          : 'bg-secondary text-muted-foreground group-hover:text-foreground group-hover:border-primary/20'
                      }`}
                    >
                      <Icon className="h-5 w-5" />
                    </span>
                    {!collapsed && (
                      <>
                        <span className="min-w-0 flex-1 truncate">{item.label}</span>
                        <span className="ml-auto text-xs opacity-0 transition-opacity group-hover:opacity-100">
                          进入
                        </span>
                      </>
                    )}
                  </Link>
                );
              })}
            </div>
          </nav>

          {/* Bottom section */}
          <div className="mt-auto border-t border-border px-6 py-3 lg:py-4">
            {!collapsed && (
              <div className="mb-4 rounded-2xl bg-secondary/60 p-3">
                <p className="text-xs tracking-[0.12em] text-muted-foreground">
                  ZAI2API · 声明
                </p>
                <p className="mt-2 text-xs text-muted-foreground">
                  本项目仅限学习与研究用途，禁止用于商业用途。
                </p>
              </div>
            )}
            <div className={`mt-4 flex items-center gap-3 ${collapsed ? 'justify-center' : ''}`}>
              {!collapsed && (
                <button
                  onClick={handleLogout}
                  className="flex-1 rounded-2xl border border-border bg-background px-4 py-3 text-sm font-medium text-muted-foreground transition-colors hover:border-destructive/40 hover:text-destructive"
                >
                  退出登录
                </button>
              )}
              <button
                className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border border-border text-muted-foreground transition-all hover:border-primary hover:text-primary"
                onClick={() => setCollapsed(!collapsed)}
                title={collapsed ? '展开侧边栏' : '收起侧边栏'}
              >
                {collapsed ? (
                  <PanelLeft className="h-4 w-4 shrink-0" />
                ) : (
                  <PanelLeftClose className="h-4 w-4 shrink-0" />
                )}
              </button>
            </div>
          </div>
        </aside>

        {/* Main content */}
        <main className="min-w-0 flex-1 overflow-hidden lg:ml-0">
          {/* Header */}
          <header className="flex min-w-0 flex-col gap-4 border-b border-border bg-card/70 px-6 py-5 backdrop-blur lg:flex-row lg:items-center lg:justify-between lg:px-10">
            <div className="flex items-center gap-3">
              <button
                className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-border text-foreground transition-colors hover:border-primary hover:text-primary lg:hidden"
                onClick={() => setSidebarOpen(true)}
                aria-label="打开导航"
              >
                <PanelLeft className="h-5 w-5" />
              </button>
              <h2 className="text-xl font-semibold text-foreground lg:text-2xl">
                {currentPageTitle}
              </h2>
            </div>
            <div className="flex flex-wrap items-center gap-3">
              <button
                onClick={() => window.location.reload()}
                className="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-primary hover:text-primary"
                title="刷新"
              >
                <RefreshCw className="inline h-4 w-4 mr-1" />
                刷新
              </button>
            </div>
          </header>

          {/* Content area */}
          <div className="h-full overflow-y-auto overflow-x-hidden bg-card/70 px-4 pb-10 pt-6 backdrop-blur lg:px-10 lg:pt-10">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
