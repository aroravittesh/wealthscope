export interface User {
  id?: string;
  email: string;
  fullName?: string;
  role?: 'USER' | 'ADMIN';
  riskPreference?: string;
  createdAt?: Date;
  updatedAt?: Date;
}

export interface Portfolio {
  id: string;
  userId?: string;
  name: string;
  description?: string;
  totalValue?: number;
  totalInvested?: number;
  totalProfitLoss?: number;
  profitLossPercentage?: number;
  createdAt?: Date;
  updatedAt?: Date;
}

export interface Holding {
  id: string;
  portfolioId: string;
  symbol: string;
  assetType: string; // stock | crypto | etf
  quantity: number;
  avgPrice: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface Asset {
  id: string;
  symbol: string;
  name: string;
  currentPrice: number;
  priceChange: number;
  priceChangePercentage: number;
  marketCap: number;
  volume: number;
  updatedAt: Date;
}

export interface Transaction {
  id: string;
  portfolioId: string;
  assetId: string;
  type: 'BUY' | 'SELL';
  quantity: number;
  price: number;
  totalAmount: number;
  fee: number;
  createdAt: Date;
}

export interface DashboardMetrics {
  totalPortfolioValue: number;
  totalInvested: number;
  totalProfitLoss: number;
  profitLossPercentage: number;
  assetsCount: number;
  portfoliosCount: number;
  topPerformers: Asset[];
  allocationData: { [key: string]: number };
}

/** Backend GET /portfolios/:id/summary (analytics engine). */
export interface AssetAllocationRow {
  symbol: string;
  assetType: string;
  costBasis: number;
  currentPrice: number;
  value: number;
  percent: number;
}

export interface PortfolioSummary {
  portfolioId: string;
  portfolioName: string;
  totalInvested: number;
  totalPortfolioValue: number;
  totalProfitLoss: number;
  profitLossPercentage: number;
  diversificationScore: number;
  volatilityScore: number;
  assetAllocation: AssetAllocationRow[];
}

export interface ChartData {
  labels: string[];
  datasets: {
    label: string;
    data: number[];
    borderColor?: string;
    backgroundColor?: string;
  }[];
}

export interface AIRecommendationPortfolioItem {
  stock: string;
}

export interface AIRecommendRequest {
  userPortfolio: AIRecommendationPortfolioItem[];
  risk?: string;
  horizon?: string;
  topN?: number;
}

export interface AIRecommendResponse {
  stock: string;
  score: number;
  decision: string;
  predictedReturn: number;
  sharpe: number;
  volatility: number;
  momentum: number;
  reason: string;
}
