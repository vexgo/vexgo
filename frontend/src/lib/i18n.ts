import { zhCN } from '@/locales/zh-CN';
import { enUS } from '@/locales/en-US';

type Translations = typeof zhCN;

const translations: Record<string, Translations> = {
  'zh-CN': zhCN,
  'en-US': enUS,
};

let currentLocale = 'zh-CN';

function detectLocale(): string {
  const savedLocale = localStorage.getItem('locale');
  if (savedLocale === 'zh-CN' || savedLocale === 'en-US') {
    return savedLocale;
  }

  if (typeof navigator !== 'undefined') {
    const languages = navigator.languages || [navigator.language];
    
    for (const lang of languages) {
      if (!lang) continue;
      
      const normalizedLang = lang.toLowerCase();
      
      if (normalizedLang.startsWith('zh')) {
        return 'zh-CN';
      }

      if (normalizedLang.startsWith('en')) {
        return 'en-US';
      }
    }
  }

  return 'zh-CN';
}

currentLocale = detectLocale();
localStorage.setItem('locale', currentLocale);

export function setLocale(locale: string) {
  if (locale === 'zh-CN' || locale === 'en-US') {
    currentLocale = locale;
    localStorage.setItem('locale', locale);
  }
}

export function getLocale(): string {
  return currentLocale;
}

export function t(key: string, params?: Record<string, unknown>): string {
  const keys = key.split('.');
  let value: unknown = translations[currentLocale];
  
  for (const k of keys) {
    if (value && typeof value === 'object' && k in value) {
      value = (value as Record<string, unknown>)[k];
    } else {
      return key;
    }
  }
  
  let result = (value as string) || key;
  
  // 如果提供了参数，进行替换
  if (params) {
    result = result.replace(/\{(\w+)\}/g, (match, paramName) => {
      return params[paramName] !== undefined ? String(params[paramName]) : match;
    });
  }
  
  return result;
}
