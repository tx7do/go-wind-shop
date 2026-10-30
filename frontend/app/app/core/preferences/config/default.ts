import type { Preferences } from "../types";

const defaultPreferences: Preferences = {
  app: {
    name: "GoWind CMS",
    title: "GoWind Shop",
    version: "0.0.0",
    locale: "zh-CN",
    isMobile: false,
    defaultAvatar: "/default-avatar.png",
    dynamicTitle: true,
    defaultPageSize: 10,
    compact: false,
  },
  copyright: {
    enable: true,
    companyName: "GoWind",
    companySiteLink: "",
    date: "2026",
    icp: "",
    icpLink: "",
  },
  logo: {
    enable: true,
    source: "/logo.svg",
  },
  theme: {
    mode: "auto",
    // 与 assets/css/main.css 的 :root 保持同一组值；这四项目前没有消费者
    // （偏好面板选色是写不读的，updateCSSVariables 也不再落 --radius），
    // 只作为 schema 的声明值存在。
    colorPrimary: "142.4 71.9% 29.2%",
    colorSuccess: "158.9 64% 30%",
    colorWarning: "38 92% 50%",
    colorDestructive: "0 73.3% 41.8%",
    radius: "0.75rem",
  },
  content: {
    hideSensitiveContent: true,
    compactMode: false,
    showRecommendations: true,
  },
  widget: {
    themeToggle: true,
    languageToggle: true,
    globalSearch: true,
    backToTop: true,
  },
  transition: {
    enable: true,
    loading: true,
    name: "fade-slide",
    progress: true,
  },
};

export { defaultPreferences };
