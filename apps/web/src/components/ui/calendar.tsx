import * as React from "react"
import { ChevronLeft, ChevronRight } from "lucide-react"
import { DayPicker } from "react-day-picker"

import { cn } from "@/lib/utils"

export type CalendarProps = React.ComponentProps<typeof DayPicker>

function Calendar({ className, classNames, components, showOutsideDays = true, ...props }: CalendarProps) {
  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      className={cn("bg-surface p-3 text-foreground", className)}
      classNames={{
        root: "w-full",
        months: "flex flex-col",
        month: "space-y-4",
        month_caption: "relative flex h-8 items-center justify-center",
        caption_label: "text-sm font-medium text-foreground",
        nav: "absolute inset-x-0 top-0 flex items-center justify-between",
        button_previous: "inline-flex h-8 w-8 items-center justify-center rounded-md text-muted hover:bg-surface-raised hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-40",
        button_next: "inline-flex h-8 w-8 items-center justify-center rounded-md text-muted hover:bg-surface-raised hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-40",
        month_grid: "w-full border-collapse space-y-1",
        weekdays: "grid grid-cols-7",
        weekday: "h-8 text-center text-[11px] font-normal uppercase text-muted",
        week: "mt-1 grid grid-cols-7",
        day: "flex h-8 w-8 items-center justify-center justify-self-center rounded-md text-sm text-foreground aria-selected:bg-accent aria-selected:text-background hover:bg-surface-raised focus-within:bg-surface-raised",
        day_button: "flex h-8 w-8 items-center justify-center rounded-md focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
        selected: "bg-accent text-background hover:bg-accent",
        today: "border border-border-strong",
        outside: "text-muted opacity-50",
        disabled: "text-muted opacity-40",
        hidden: "invisible",
        ...classNames
      }}
      components={{
        Chevron: ({ className: chevronClassName, orientation }) =>
          orientation === "left" ? (
            <ChevronLeft className={cn("h-4 w-4", chevronClassName)} />
          ) : (
            <ChevronRight className={cn("h-4 w-4", chevronClassName)} />
          ),
        ...components
      }}
      {...props}
    />
  )
}
Calendar.displayName = "Calendar"

export { Calendar }
