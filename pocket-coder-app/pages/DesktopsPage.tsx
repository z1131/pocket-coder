import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { DesktopItem, fetchDesktops, renameDesktop } from '../api/client';
import { useAuth } from '../hooks/useAuth';
import { usePocketWS } from '../hooks/usePocketWS';
import { ConnectionStatus } from '../types';

const StatusDot: React.FC<{ status: string }> = ({ status }) => {
  const color = status === 'online' ? 'bg-emerald-500' : status === 'busy' ? 'bg-amber-500' : 'bg-neutral-500';
  return <span className={`inline-block w-2 h-2 rounded-full ${color} mr-2`} />;
};

const DesktopsPage: React.FC = () => {
  const { accessToken, clear } = useAuth();
  const navigate = useNavigate();
  const [list, setList] = useState<DesktopItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const { status: wsStatus, presence } = usePocketWS({ token: accessToken });

  const statusText = useMemo(() => {
    if (wsStatus === ConnectionStatus.CONNECTING) return 'WS：连接中...';
    if (wsStatus === ConnectionStatus.BUSY) return 'WS：忙碌 (agent running)';
    if (wsStatus === ConnectionStatus.CONNECTED) return 'WS：已连接';
    return 'WS：未连接';
  }, [wsStatus]);

  useEffect(() => {
    if (!accessToken) {
      navigate('/login', { replace: true });
      return;
    }
    const load = async () => {
      try {
        setLoading(true);
        const res = await fetchDesktops(accessToken);
        setList(res.desktops || []);
      } catch (err: any) {
        setError(err?.message || '加载失败');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [accessToken, navigate]);

  const handleRename = async (id: number, current: string) => {
    const name = window.prompt('输入新的设备名称', current);
    if (!name || name === current) return;
    try {
      await renameDesktop(id, name, accessToken);
      setList((prev) => prev.map((d) => (d.id === id ? { ...d, name } : d)));
    } catch (err: any) {
      window.alert(err?.message || '重命名失败');
    }
  };

  const handleLogout = () => {
    clear();
    navigate('/login', { replace: true });
  };

  const handleEnter = (id: number) => {
    navigate(`/desktops/${id}/sessions`);
  };

  const getStatus = (desktop: DesktopItem) => {
    const ws = presence[desktop.id];
    return ws || desktop.status || 'offline';
  };

  return (
    <div className="min-h-screen bg-black text-white">
      <header className="border-b border-neutral-900 px-6 py-4 flex items-center justify-between">
        <div>
          <div className="text-sm text-neutral-400">Pocket Coder</div>
          <div className="text-lg font-semibold">我的电脑</div>
        </div>
        <div className="flex items-center gap-3 text-sm">
          <span className="text-neutral-400">{statusText}</span>
          <button onClick={handleLogout} className="px-3 py-2 rounded-lg bg-neutral-800 hover:bg-neutral-700 border border-neutral-700">退出</button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-8">
        {loading && <div className="text-neutral-400">加载中...</div>}
        {error && <div className="text-rose-400 mb-4">{error}</div>}

        {!loading && !error && (
          <div className="grid gap-4">
            {list.length === 0 && <div className="text-neutral-400">暂无设备。请在电脑端登录并绑定。</div>}
            {list.map((d) => (
              <div key={d.id} className="border border-neutral-800 rounded-xl p-4 bg-neutral-900/60 flex items-center justify-between">
                <div className="flex flex-col gap-1">
                  <div className="text-lg font-semibold flex items-center">
                    <StatusDot status={getStatus(d)} />
                    {d.name}
                  </div>
                  <div className="text-sm text-neutral-400">ID: {d.id} · Agent: {d.agent_type}</div>
                  {d.os_info && <div className="text-xs text-neutral-500">OS: {d.os_info}</div>}
                  {d.working_dir && <div className="text-xs text-neutral-500">工作目录: {d.working_dir}</div>}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleEnter(d.id)}
                    className="px-3 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-sm text-white"
                  >进入会话</button>
                  <button
                    onClick={() => handleRename(d.id, d.name)}
                    className="px-3 py-2 rounded-lg border border-neutral-700 bg-neutral-800 hover:bg-neutral-700 text-sm"
                  >重命名</button>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
};

export default DesktopsPage;
