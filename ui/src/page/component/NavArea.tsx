export function NavArea({ handlePrev, handleNext }: { handlePrev: () => void, handleNext: () => void }) {
    return (
      <div className="nav-area">
        <button onClick={() => handlePrev()} className="button">
          Back
        </button>
        <button onClick={() => handleNext()} className="button">
          Next
        </button>
      </div>
    )
  }
  