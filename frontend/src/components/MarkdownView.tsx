import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeSanitize from 'rehype-sanitize'

// Project/section bodies are Markdown; render + sanitize on the client with a
// vetted allowlist (no scripts/iframes). Never trust stored Markdown as safe HTML.
export function MarkdownView({ children }: { children: string }) {
  return (
    <div className="markdown" style={{ lineHeight: 1.6, color: 'var(--ink)' }}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>
        {children}
      </ReactMarkdown>
    </div>
  )
}
