import type { FormEvent } from 'react'
import type { UseTagsInputResult } from '../hooks/useTagsInput'

type PostFormProps = {
  heading: string
  submitLabel: string
  title: string
  content: string
  contentType?: 0 | 1
  onContentTypeChange?: (value: 0 | 1) => void
  onTitleChange: (value: string) => void
  onContentChange: (value: string) => void
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
  onCancel: () => void
  tagsInput: UseTagsInputResult
  tagPrefix?: string
  tagClassName?: string
  tagRemoveClassName?: string
}

export default function PostForm({
  heading,
  submitLabel,
  title,
  content,
  contentType,
  onContentTypeChange,
  onTitleChange,
  onContentChange,
  onSubmit,
  onCancel,
  tagsInput,
  tagPrefix = '',
  tagClassName = 'inline-flex items-center px-3 py-1 rounded-full bg-primary-50 text-primary-700 text-sm border border-primary-100',
  tagRemoveClassName = 'ml-2 text-primary-500 hover:text-primary-700',
}: PostFormProps) {
  const {
    tags,
    tagInput,
    tagError,
    addTag,
    removeTag,
    handleTagInputChange,
    handleCompositionStart,
    handleCompositionEnd,
    handleTagKeyDown,
  } = tagsInput
  const showContentType = typeof contentType === 'number' && typeof onContentTypeChange === 'function'
  const contentHint = showContentType
    ? contentType === 1
      ? 'Markdown 文本将按语法渲染'
      : '普通文本将按原样展示'
    : '支持 Markdown 格式'

  return (
    <form onSubmit={onSubmit} className="card space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">{heading}</h1>

      <div>
        <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
          标题
        </label>
        <input
          id="title"
          type="text"
          value={title}
          onChange={(e) => onTitleChange(e.target.value)}
          placeholder="输入帖子标题..."
          required
          className="input"
        />
      </div>

      <div>
        <label htmlFor="tags" className="block text-sm font-medium text-gray-700 mb-2">
          标签
        </label>
        <div className="flex flex-col space-y-3">
          <div className="flex space-x-3">
            <input
              id="tags"
              type="text"
              value={tagInput}
              onChange={(e) => handleTagInputChange(e.target.value)}
              onCompositionStart={handleCompositionStart}
              onCompositionEnd={handleCompositionEnd}
              onKeyDown={handleTagKeyDown}
              placeholder="输入标签，按回车或点击添加"
              maxLength={6}
              className="input"
            />
            <button
              type="button"
              onClick={addTag}
              className="btn-secondary whitespace-nowrap"
              disabled={tags.length >= 4}
            >
              添加标签
            </button>
          </div>
          <div className="flex flex-wrap gap-2">
            {tags.map(tag => (
              <span key={tag} className={tagClassName}>
                {tagPrefix}{tag}
                <button
                  type="button"
                  onClick={() => removeTag(tag)}
                  className={tagRemoveClassName}
                  aria-label={`移除标签 ${tag}`}
                >
                  &times;
                </button>
              </span>
            ))}
            {tags.length === 0 && (
              <span className="text-sm text-gray-500">最多 4 个标签，每个不超过 6 个字</span>
            )}
          </div>
          {tagError && <p className="text-sm text-red-500">{tagError}</p>}
        </div>
      </div>

      {showContentType && (
        <div>
          <span className="block text-sm font-medium text-gray-700 mb-2">文本格式</span>
          <div className="grid gap-3 sm:grid-cols-2">
            <label
              className={`flex cursor-pointer items-start gap-3 rounded-xl border px-4 py-3 text-sm shadow-sm transition ${
                contentType === 0
                  ? 'border-primary-300 bg-primary-50/60'
                  : 'border-gray-200/70 bg-white/70 hover:border-gray-300'
              }`}
            >
              <input
                type="radio"
                name="contentType"
                value="0"
                checked={contentType === 0}
                onChange={() => onContentTypeChange?.(0)}
                className="mt-1 h-4 w-4 text-primary-600 focus:ring-primary-200"
              />
              <div>
                <p className="font-medium text-gray-900">普通文本</p>
                <p className="text-xs text-gray-500">不解析 Markdown 语法</p>
              </div>
            </label>
            <label
              className={`flex cursor-pointer items-start gap-3 rounded-xl border px-4 py-3 text-sm shadow-sm transition ${
                contentType === 1
                  ? 'border-primary-300 bg-primary-50/60'
                  : 'border-gray-200/70 bg-white/70 hover:border-gray-300'
              }`}
            >
              <input
                type="radio"
                name="contentType"
                value="1"
                checked={contentType === 1}
                onChange={() => onContentTypeChange?.(1)}
                className="mt-1 h-4 w-4 text-primary-600 focus:ring-primary-200"
              />
              <div>
                <p className="font-medium text-gray-900">Markdown 文本</p>
                <p className="text-xs text-gray-500">支持标题、列表、代码块</p>
              </div>
            </label>
          </div>
        </div>
      )}

      <div>
        <label htmlFor="content" className="block text-sm font-medium text-gray-700 mb-2">
          内容
        </label>
        <textarea
          id="content"
          value={content}
          onChange={(e) => onContentChange(e.target.value)}
          placeholder="输入帖子内容..."
          rows={12}
          required
          className="textarea"
        />
        <p className="mt-2 text-sm text-gray-500">{contentHint}</p>
      </div>

      <div className="flex justify-end space-x-4 pt-4 border-t border-gray-200">
        <button type="button" onClick={onCancel} className="btn-secondary">
          取消
        </button>
        <button type="submit" className="btn-primary">
          {submitLabel}
        </button>
      </div>
    </form>
  )
}
