import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from 'sonner'

import '@fontsource/baloo-2/500.css'
import '@fontsource/baloo-2/600.css'
import '@fontsource/baloo-2/700.css'
import '@fontsource/baloo-2/800.css'
import '@fontsource/hanken-grotesk/400.css'
import '@fontsource/hanken-grotesk/500.css'
import '@fontsource/hanken-grotesk/600.css'
import '@fontsource/hanken-grotesk/700.css'
import '@fontsource/space-mono/400.css'
import '@fontsource/space-mono/700.css'

import './index.css'
import App from './App'
import { queryClient } from './api/queryClient'
import { ThemeProvider } from './lib/theme'
import { SoundProvider } from './lib/sound'
import { LangProvider } from './i18n/lang'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <LangProvider>
          <SoundProvider>
            <BrowserRouter>
              <App />
              <Toaster position="top-center" richColors closeButton />
            </BrowserRouter>
          </SoundProvider>
        </LangProvider>
      </ThemeProvider>
    </QueryClientProvider>
  </StrictMode>,
)
