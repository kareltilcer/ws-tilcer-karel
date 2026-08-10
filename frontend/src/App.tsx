import { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import { SiteLayout } from './components/SiteLayout'
import { Spinner } from './components/states'
import Home from './pages/Home'
import Projects from './pages/Projects'
import ProjectDetail from './pages/ProjectDetail'
import Linktree from './pages/Linktree'
import About from './pages/About'
import Arcade from './pages/Arcade'
import Contact from './pages/Contact'
import Legal from './pages/Legal'
import NotFound from './pages/NotFound'

// The admin CMS is a lazy chunk kept out of the public bundle (PRD §7).
const AdminApp = lazy(() => import('./admin/AdminApp'))
// Skills is lazy so its ~90 KB generated icon registry stays out of the main
// bundle — only visitors to /skills (and the lazy admin) pay for it.
const Skills = lazy(() => import('./pages/Skills'))

function PageFallback() {
  return (
    <div style={{ display: 'grid', placeItems: 'center', minHeight: '60vh' }}>
      <Spinner size={36} />
    </div>
  )
}

export default function App() {
  return (
    <Routes>
      <Route element={<SiteLayout />}>
        <Route path="/" element={<Home />} />
        <Route path="/projects" element={<Projects />} />
        <Route path="/projects/:slug" element={<ProjectDetail />} />
        <Route path="/links" element={<Linktree />} />
        <Route path="/about" element={<About />} />
        <Route
          path="/skills"
          element={
            <Suspense fallback={<PageFallback />}>
              <Skills />
            </Suspense>
          }
        />
        <Route path="/arcade" element={<Arcade />} />
        <Route path="/contact" element={<Contact />} />
        <Route path="/legal" element={<Legal />} />
        <Route path="*" element={<NotFound />} />
      </Route>
      <Route
        path="/admin/*"
        element={
          <Suspense fallback={<PageFallback />}>
            <AdminApp />
          </Suspense>
        }
      />
    </Routes>
  )
}
