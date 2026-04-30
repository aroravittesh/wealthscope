import torch
import torch.nn as nn
import pandas as pd
import numpy as np
import yfinance as yf
import joblib
import ta
import warnings

warnings.filterwarnings("ignore")

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")


# ======================
# MODEL DEFINITION (UNCHANGED)
# ======================
class Transformer(nn.Module):
    def __init__(self):
        super().__init__()
        self.embed = nn.Linear(7, 64)
        layer = nn.TransformerEncoderLayer(
            d_model=64,
            nhead=4,
            batch_first=True,
            dropout=0.1,
        )
        self.trans = nn.TransformerEncoder(layer, num_layers=4)
        self.fc = nn.Linear(64, 1)

    def forward(self, x):
        x = self.embed(x)
        x = self.trans(x)
        return self.fc(x[:, -1, :])


# ======================
# LOAD MODEL
# ======================
def load_model(model_path="final_model.pth", scaler_path="scaler.pkl"):
    model = Transformer().to(device)
    state_dict = torch.load(model_path, map_location=device)
    model.load_state_dict(state_dict)
    model.eval()

    scaler = joblib.load(scaler_path)
    return model, scaler


model, scaler = load_model()


# ======================
# FEATURE ENGINEERING
# ======================
def get_latest_features(stock, seq_len=60):
    df = yf.download(stock, period="1y", progress=False, auto_adjust=False)

    if df.empty or len(df) < seq_len + 10:
        return None

    close = df["Close"].squeeze()

    df["returns"] = close.pct_change()
    df["volatility"] = df["returns"].rolling(10).std()
    df["rsi"] = ta.momentum.RSIIndicator(close=close).rsi()
    df["ma20"] = close.rolling(20).mean()
    df["ema"] = close.ewm(span=20).mean()
    df["macd"] = df["ema"] - close.ewm(span=50).mean()
    df["price_change"] = close.diff()

    df = df.replace([np.inf, -np.inf], np.nan).dropna()

    if len(df) < seq_len:
        return None

    features = df[
        ["Close", "Volume", "rsi", "volatility", "ema", "macd", "price_change"]
    ].iloc[-seq_len:]

    scaled = scaler.transform(features)

    return torch.tensor(scaled, dtype=torch.float32).unsqueeze(0).to(device)


# ======================
# HELPER FUNCTIONS
# ======================
def sharpe_ratio(returns):
    return np.mean(returns) / (np.std(returns) + 1e-6)


# ======================
# MAIN RECOMMEND FUNCTION (OPTIMIZED)
# ======================
def recommend(user_portfolio, risk="medium", horizon="long", top_n=5):

    owned = {p["stock"] for p in user_portfolio}
    recs = []

    stock_list = [
        "AAPL", "MSFT", "GOOG", "AMZN", "TSLA", "NVDA", "META",
        "NFLX", "AMD", "INTC", "BABA", "ORCL", "IBM", "ADBE",
        "CSCO", "QCOM", "SHOP",
    ]

    for s in stock_list:

        if s in owned:
            continue

        try:
            X = get_latest_features(s)

            if X is None:
                continue

            with torch.no_grad():
                pred_scaled = model(X).item()

            df = yf.download(s, period="2y", progress=False)

            returns = df["Close"].squeeze().pct_change().dropna()

            if len(returns) == 0:
                continue

            sharpe = float(sharpe_ratio(returns))
            vol = float(returns.tail(20).std())

            momentum = float(
                df["Close"].squeeze().iloc[-1] /
                df["Close"].squeeze().rolling(20).mean().iloc[-1]
            )

            # ======================
            # 🔥 OPTIMIZED SCORING
            # ======================
            norm_momentum = np.tanh(momentum)
            norm_sharpe = np.tanh(sharpe)
            norm_pred = np.tanh(pred_scaled)

            score = 0.4 * norm_momentum + 0.35 * norm_sharpe + 0.25 * norm_pred

            # Risk adjustment
            if risk == "low":
                score -= 1.2 * vol
            elif risk == "high":
                score += 0.6 * vol

            # Horizon adjustment
            if horizon == "long":
                score += 0.4 * norm_sharpe
            elif horizon == "short":
                score += 0.3 * norm_momentum

            # ======================
            # 🔥 IMPROVED DECISION
            # ======================
            if score > 0.6:
                decision = "BUY"
            elif score > 0.35:
                decision = "HOLD"
            else:
                decision = "AVOID"

            recs.append(
                {
                    "stock": s,
                    "score": round(score, 4),
                    "decision": decision,
                    "predicted_return": round(pred_scaled, 4),
                    "sharpe": round(sharpe, 3),
                    "volatility": round(vol, 4),
                    "momentum": round(momentum, 3),
                    "reason": (
                        f"Prediction: {pred_scaled:.3f}, "
                        f"Momentum: {momentum:.2f}, "
                        f"Sharpe: {sharpe:.2f}, "
                        f"Volatility: {vol:.3f}"
                    ),
                }
            )

        except Exception:
            continue

    recs = sorted(recs, key=lambda x: x["score"], reverse=True)[:top_n]

    return recs
