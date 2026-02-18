import { createContext, useContext } from 'react';

export type Lang = 'ja' | 'en';

const translations = {
  ja: {
    appTitle: 'File Bridge',
    serverRunning: 'サーバー起動中',
    serverStopped: 'サーバー停止中',
    saveLocation: '保存先',
    notSet: '未設定',
    change: '変更',
    recentUploads: '受信履歴',
    noFiles: 'まだファイルを受信していません',
    firewallWarning:
      'iPhoneから接続できない場合は、Windowsファイアウォールの設定を確認してください。このアプリの通信を許可するか、一時的にファイアウォールを無効にしてください。両方のデバイスが同じWi-Fiに接続されている必要があります。',
  },
  en: {
    appTitle: 'File Bridge',
    serverRunning: 'Server Running',
    serverStopped: 'Server Stopped',
    saveLocation: 'Save Location',
    notSet: 'Not set',
    change: 'Change',
    recentUploads: 'Recent Uploads',
    noFiles: 'No files received yet',
    firewallWarning:
      'If your iPhone cannot connect, check Windows Firewall settings. Allow this app through the firewall or temporarily disable it. Both devices must be on the same Wi-Fi network.',
  },
} as const;

export type TranslationKey = keyof (typeof translations)['ja'];

export interface LanguageContextType {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: TranslationKey) => string;
}

export const LanguageContext = createContext<LanguageContextType>({
  lang: 'ja',
  setLang: () => {},
  t: (key) => translations.ja[key],
});

export function useLanguage() {
  return useContext(LanguageContext);
}

export function getTranslation(lang: Lang, key: TranslationKey): string {
  return translations[lang]?.[key] ?? translations.ja[key];
}
