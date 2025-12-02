/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
}

interface ImportMeta {
  readonly any: ImportMetaEnv;
}

declare module 'vite' {
  interface ImportMeta {
    readonly any: ImportMetaEnv;
  }
}
