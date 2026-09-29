export interface JogoDetalheReviewProps {
  review?: string | null
  onAdicionarReview: () => void
}

export function JogoDetalheReview({ review, onAdicionarReview }: JogoDetalheReviewProps) {
  const temReview = Boolean(review && review.trim())

  return (
    <div className="flex flex-col gap-2.5">
      <h2 className="text-[15px] sm:text-[16px] font-semibold text-[var(--text-primary)]">
        Review
      </h2>
      <div className="rounded-[14px] border border-[var(--detalhe-bloco-border)] bg-[var(--detalhe-bloco-bg)] p-5 sm:p-[20px_22px]">
        {temReview ? (
          <p className="whitespace-pre-line text-[15px] leading-[1.65] text-[var(--detalhe-text-review)]">
            {review}
          </p>
        ) : (
          <div className="flex items-center gap-2 text-[15px] text-[var(--detalhe-text-meta)]">
            <span>Sem review.</span>
            <button
              type="button"
              onClick={onAdicionarReview}
              className="font-medium text-[var(--accent)] hover:underline focus:outline-none"
            >
              Adicionar review
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
