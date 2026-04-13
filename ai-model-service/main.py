from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import uvicorn
from typing import List, Optional
from predict import recommend as recommend_stocks

app = FastAPI(title="Stock ML Recommendation Service")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
def health():
    return {"status": "ok"}


class PortfolioItem(BaseModel):
    stock: str


class RecommendationRequest(BaseModel):
    user_portfolio: List[PortfolioItem]
    risk: Optional[str] = "medium"
    horizon: Optional[str] = "long"
    top_n: Optional[int] = 5


class RecommendationResponse(BaseModel):
    stock: str
    score: float
    decision: str
    predicted_return: float
    sharpe: float
    volatility: float
    momentum: float
    reason: str


@app.post("/recommend", response_model=List[RecommendationResponse])
def recommend_endpoint(req: RecommendationRequest):
    try:
        recs = recommend_stocks(
            user_portfolio=[{"stock": item.stock} for item in req.user_portfolio],
            risk=req.risk,
            horizon=req.horizon,
            top_n=req.top_n,
        )
        return recs
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
