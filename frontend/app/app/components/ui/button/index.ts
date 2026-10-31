import { type VariantProps, cva } from "class-variance-authority"

export { default as Button } from "./Button.vue"

export const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-medium transition-colors duration-200 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
  {
    variants: {
      // A 稿：填充按钮只靠明度与描边分层，不加彩色投影光晕（原 hover:shadow-[…primary/0.2] 是赛博风残留）
      // 禁用态不再用 opacity-50：半透明的填充按钮会同时稀释底和字，
      // 亮色下白字落淡绿约 1.5:1、暗色下深墨落深绿约 2:1，两态都读不出来。
      // 改成中性的 muted 底 + muted 字（亮 4.6:1 / 暗 5.3:1），一眼是「不可用」且仍可读。
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90 disabled:bg-muted disabled:text-muted-foreground",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90 disabled:bg-muted disabled:text-muted-foreground",
        outline: "border border-input bg-card text-foreground hover:border-primary/60 hover:bg-muted disabled:border-border disabled:bg-transparent disabled:text-muted-foreground",
        secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80 disabled:bg-muted disabled:text-muted-foreground",
        ghost: "hover:bg-muted disabled:text-muted-foreground",
        link: "text-primary underline-offset-4 hover:underline disabled:text-muted-foreground",
      },
      size: {
        default: "h-10 px-5 py-2",
        sm: "h-8 px-3 text-xs",
        lg: "h-11 px-8",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)

export type ButtonVariants = VariantProps<typeof buttonVariants>
