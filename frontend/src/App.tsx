import { useState, useEffect, useCallback } from 'react';
import { QRCodeSVG } from 'qrcode.react';
import './App.css';
import { GetServerInfo, SelectSaveDir, GetUploadHistory, SetLang } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { LanguageContext, getTranslation, type Lang, type TranslationKey } from './i18n';

interface ServerInfo {
  running: boolean;
  url: string;
  port: number;
  lanIP: string;
  saveDir: string;
  lang: string;
}

interface UploadRecord {
  fileName: string;
  size: number;
  timestamp: string;
  savePath: string;
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
}

function App() {
  const [serverInfo, setServerInfo] = useState<ServerInfo | null>(null);
  const [history, setHistory] = useState<UploadRecord[]>([]);
  const [lang, setLangState] = useState<Lang>('ja');

  const t = useCallback((key: TranslationKey) => getTranslation(lang, key), [lang]);

  const refreshInfo = useCallback(async () => {
    try {
      const info = await GetServerInfo();
      const si = info as unknown as ServerInfo;
      setServerInfo(si);
      if (si.lang === 'ja' || si.lang === 'en') {
        setLangState(si.lang as Lang);
      }
    } catch (e) {
      console.error('Failed to get server info:', e);
    }
  }, []);

  const refreshHistory = useCallback(async () => {
    try {
      const h = await GetUploadHistory();
      setHistory((h || []) as unknown as UploadRecord[]);
    } catch (e) {
      console.error('Failed to get history:', e);
    }
  }, []);

  useEffect(() => {
    refreshInfo();
    refreshHistory();

    const cancel = EventsOn('upload:completed', () => {
      refreshHistory();
    });

    const interval = setInterval(refreshInfo, 5000);

    return () => {
      cancel();
      clearInterval(interval);
    };
  }, [refreshInfo, refreshHistory]);

  const handleSetLang = async (newLang: Lang) => {
    setLangState(newLang);
    try {
      await SetLang(newLang);
      // Refresh to get updated URL with new lang param
      await refreshInfo();
    } catch (e) {
      console.error('Failed to set language:', e);
    }
  };

  const handleSelectDir = async () => {
    try {
      const newDir = await SelectSaveDir();
      if (newDir) {
        setServerInfo(prev => prev ? { ...prev, saveDir: newDir } : null);
      }
    } catch (e) {
      console.error('Failed to select directory:', e);
    }
  };

  const running = serverInfo?.running ?? false;
  const url = serverInfo?.url ?? '';

  return (
    <LanguageContext.Provider value={{ lang, setLang: handleSetLang, t }}>
      <div id="app">
        <div className="header">
          <div className="lang-switcher">
            <button
              className={`lang-btn ${lang === 'ja' ? 'active' : ''}`}
              onClick={() => handleSetLang('ja')}
            >
              JA
            </button>
            <button
              className={`lang-btn ${lang === 'en' ? 'active' : ''}`}
              onClick={() => handleSetLang('en')}
            >
              EN
            </button>
          </div>
          <h1>{t('appTitle')}</h1>
          <p>{t('appSubtitle')}</p>
        </div>

        <div style={{ textAlign: 'center' }}>
          <span className={`status-badge ${running ? 'running' : 'stopped'}`}>
            <span className="status-dot" />
            {running ? t('serverRunning') : t('serverStopped')}
          </span>
        </div>

        {running && url && (
          <div className="qr-section">
            <div className="qr-wrapper">
              <QRCodeSVG value={url} size={180} level="M" />
            </div>
            <div className="url-text">{url}</div>
            <div className="port-text">{t('port')}{serverInfo?.port}</div>
          </div>
        )}

        <div className="save-section">
          <div className="label">{t('saveLocation')}</div>
          <div className="dir-row">
            <div className="dir-path" title={serverInfo?.saveDir}>
              {serverInfo?.saveDir ?? t('notSet')}
            </div>
            <button className="change-btn" onClick={handleSelectDir}>
              {t('change')}
            </button>
          </div>
        </div>

        <div className="history-section">
          <div className="label">{t('recentUploads')}</div>
          {history.length === 0 ? (
            <div className="empty-history">{t('noFiles')}</div>
          ) : (
            history.map((record, i) => (
              <div key={i} className="history-item">
                <div className="file-info">
                  <div className="file-name" title={record.fileName}>
                    {record.fileName}
                  </div>
                  <div className="file-meta">{record.timestamp}</div>
                </div>
                <div className="file-size">{formatSize(record.size)}</div>
              </div>
            ))
          )}
        </div>

        <div className="warning-box">
          {t('firewallWarning')}
        </div>
      </div>
    </LanguageContext.Provider>
  );
}

export default App;
