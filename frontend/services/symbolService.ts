import api from './api';

export interface Symbol {
  symbol: string;
  base_asset: string;
  quote_asset: string;
}

export interface SymbolsResponse {
  total: number;
  symbols: Symbol[];
}

// Get Binance Spot symbols
export const getBinanceSpotSymbols = async (): Promise<Symbol[]> => {
  const response = await api.get<SymbolsResponse>('/binance/spot/symbols');
  return response.data.symbols;
};

// Get Binance Futures symbols
export const getBinanceFuturesSymbols = async (): Promise<Symbol[]> => {
  const response = await api.get<SymbolsResponse>('/binance/futures/symbols');
  return response.data.symbols;
};

// Get Bittrex symbols
export const getBittrexSymbols = async (): Promise<Symbol[]> => {
  const response = await api.get<SymbolsResponse>('/bittrex/symbols');
  return response.data.symbols;
};

// Get symbols based on exchange and trading mode
export const getSymbols = async (
  exchange: string,
  tradingMode: string,
): Promise<Symbol[]> => {
  if (exchange === 'binance') {
    if (tradingMode === 'spot') {
      return getBinanceSpotSymbols();
    } else if (tradingMode === 'futures' || tradingMode === 'margin') {
      return getBinanceFuturesSymbols();
    }
  } else if (exchange === 'bittrex') {
    return getBittrexSymbols();
  }
  return [];
};
