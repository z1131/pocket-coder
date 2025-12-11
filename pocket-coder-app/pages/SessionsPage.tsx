import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { fetchSessions, SessionItem, fetchDesktops, DesktopItem, createSession } from '../api/client';
import { useAuth } from '../hooks/useAuth';

const SessionsPage: React.FC = () => {
  const { accessToken } = useAuth();
  const navigate = useNavigate();
  const { desktopId } = useParams<{ desktopId: string }>();
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [desktop, setDesktop] = useState<DesktopItem | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!accessToken) {
      navigate('/login', { replace: true });
      return;
    }

    const id = Number(desktopId);
    if (!id) {
      navigate('/desktops', { replace: true });
      return;
    }

    const load = async () => {
      try {
        setLoading(true);
        // 获取设备信息
        const desktopsRes = await fetchDesktops(accessToken);
        const found = desktopsRes.desktops.find((d) => d.id === id);
        if (found) setDesktop(found);

        // 获取会话列表
        const res = await fetchSessions(id, accessToken);
        setSessions(res.sessions || []);
      } catch (err: any) {
        setError(err?.message || '加载失败');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [accessToken, desktopId, navigate]);

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const handleEnterSession = (sessionId: number) => {
    navigate(`/desktops/${desktopId}/sessions/${sessionId}`);
  };

  const handleNewSession = async () => {
    if (!accessToken || !desktopId) return;
    try {
      setLoading(true);
      const newSession = await createSession(Number(desktopId), accessToken);
      navigate(`/desktops/${desktopId}/sessions/${newSession.id}`);
    } catch (err: any) {
      setError(err?.message || '创建会话失败');
      setLoading(false);
    }
  };

  const activeSessions = sessions.filter(s => s.status === 'active');
  const historySessions = sessions.filter(s => s.status !== 'active');

  const renderSessionCard = (s: SessionItem) => (
    <div
      key={s.id}
      onClick={() => handleEnterSession(s.id)}
      className={`border rounded-xl p-4 cursor-pointer transition-colors ${
        s.status === 'active' 
          ? 'border-emerald-800 bg-emerald-900/10 hover:bg-emerald-900/20' 
          : 'border-neutral-800 bg-neutral-900/60 hover:bg-neutral-800/60'
      }`}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="text-lg font-semibold truncate flex items-center gap-2">
            {s.title || `终端 #${s.id}`}
            {s.status === 'active' && (
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            )}
          </div>
          {s.summary && (
            <div className="text-sm text-neutral-400 mt-1 line-clamp-2">
              {s.summary}
            </div>
          )}
          <div className="text-xs text-neutral-500 mt-2">
            {formatDate(s.started_at)}
            {s.status === 'ended' && ' · 已结束'}
          </div>
        </div>
        <div className="ml-4 flex-shrink-0">
          <span
            className={`inline-block px-2 py-1 rounded text-xs ${
              s.status === 'active'
                ? 'bg-emerald-900/50 text-emerald-400'
                : 'bg-neutral-800 text-neutral-500'
            }`}
          >
            {s.status === 'active' ? '运行中' : '已归档'}
          </span>
        </div>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-black text-white">
      <header className="border-b border-neutral-900 px-6 py-4 flex items-center justify-between">
        <div>
          <button
            onClick={() => navigate('/desktops')}
            className="text-sm text-neutral-400 hover:text-white mb-1"
          >
            &larr; 返回设备列表
          </button>
          <div className="text-lg font-semibold">{desktop?.name || '设备'} 的终端会话</div>
        </div>
        <button
          onClick={handleNewSession}
          disabled={loading}
          className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-sm text-white disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? '处理中...' : '+ 新建终端'}
        </button>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-8">
        {error && <div className="text-rose-400 mb-4 bg-rose-900/20 p-3 rounded">{error}</div>}

        {!loading && (
          <div className="space-y-8">
            {/* Active Sessions */}
            <section>
              <h2 className="text-sm font-bold text-neutral-400 uppercase tracking-wider mb-4 flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
                活跃终端 ({activeSessions.length})
              </h2>
              <div className="grid gap-4">
                {activeSessions.length === 0 && (
                  <div className="text-neutral-500 italic text-sm py-4 border border-dashed border-neutral-800 rounded-lg text-center">
                    当前没有运行中的终端
                  </div>
                )}
                {activeSessions.map(renderSessionCard)}
              </div>
            </section>

            {/* History Sessions */}
            <section>
              <h2 className="text-sm font-bold text-neutral-400 uppercase tracking-wider mb-4">
                历史记录 ({historySessions.length})
              </h2>
              <div className="grid gap-4">
                {historySessions.length === 0 && (
                  <div className="text-neutral-500 italic text-sm">暂无历史记录</div>
                )}
                {historySessions.map(renderSessionCard)}
              </div>
            </section>
          </div>
        )}
      </main>
    </div>
  );
};

export default SessionsPage;
