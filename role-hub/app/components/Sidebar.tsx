import { CATEGORIES, FRAMEWORKS } from '~/constants/filters'

interface SidebarProps {
  selectedCategory: string
  onCategoryChange: (category: string) => void
  selectedFrameworks: string[]
  onFrameworkToggle: (framework: string) => void
}

export function Sidebar({
  selectedCategory,
  onCategoryChange,
  selectedFrameworks,
  onFrameworkToggle,
}: SidebarProps) {
  return (
    <aside className="flex flex-col gap-8 w-60 flex-shrink-0">
      <div className="flex flex-col gap-4">
        <h4 className="text-xs font-bold text-text-sub tracking-widest uppercase font-ui m-0">Categories</h4>
        {CATEGORIES.map((cat) => (
          <button
            key={cat}
            onClick={() => onCategoryChange(cat)}
            className={`text-left text-sm font-medium font-ui bg-transparent border-none cursor-pointer transition-colors ${
              selectedCategory === cat ? 'text-primary font-semibold' : 'text-text-sub hover:text-text-main'
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      <div className="flex flex-col gap-4">
        <h4 className="text-xs font-bold text-text-sub tracking-widest uppercase font-ui m-0">Frameworks</h4>
        {FRAMEWORKS.map((fw) => (
          <label key={fw} className="flex items-center gap-3 text-sm text-text-sub font-ui cursor-pointer hover:text-text-main transition-colors">
            <input
              type="checkbox"
              checked={selectedFrameworks.includes(fw)}
              onChange={() => onFrameworkToggle(fw)}
              className="accent-primary"
            />
            {fw}
          </label>
        ))}
      </div>
    </aside>
  )
}
