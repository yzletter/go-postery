import type { FormEvent } from 'react'
import type { UseTagsInputResult } from '../hooks/useTagsInput'

type PostFormProps = {
  heading: string
  submitLabel: string
  title: string
  content: string
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
        <p className="mt-2 text-sm text-gray-500">支持 Markdown 格式</p>
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
