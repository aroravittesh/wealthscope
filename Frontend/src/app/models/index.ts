export interface User {
  id: string;
  email: string;
  fullName: string;
  role: 'USER' | 'ADMIN';
  createdAt: Date;
  updatedAt: Date;
}

export interface AuthResponse {
  user: User;
  token: string;
  refreshToken: string;
}

export interface Portfolio {
  id: string;
  userId: string;
  name: string;
  description: string;
  totalValue: number;
  totalInvested: number;
  totalProfitLoss: number;
  profitLossPercentage: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface Holding {
  id: string;
  portfolioId: string;
  assetId: string;
  quantity: number;
  purchasePrice: number;
  currentPrice: number;
  totalCost: number;
  currentValue: number;
  profitLoss: number;
  profitLossPercentage: number;
  purchasedAt: Date;
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

export interface ChartData {
  labels: string[];
  datasets: {
    label: string;
    data: number[];
    borderColor?: string;
    backgroundColor?: string;
  }[];
}
