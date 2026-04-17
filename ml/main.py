import sys
import os
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

# Ensure submodules can import their own files
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from premium.main import router as premium_router, load_model_instance as premium_load
from fraud.main import router as fraud_router
from forecast.main import router as forecast_router, train_all_zones as forecast_train

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("unified-ml")

@asynccontextmanager
async def lifespan(app: FastAPI):
    log.info("Starting Unified ML Service...")
    premium_load()
    log.info("Premium models loaded.")
    forecast_train()
    log.info("Forecast models trained.")
    yield
    log.info("Shutting down Unified ML Service.")

app = FastAPI(title="InDel Unified ML Service", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(premium_router)
app.include_router(fraud_router)
app.include_router(forecast_router)

@app.get("/health")
def health():
    return {
        "status": "ok",
        "service": "unified-ml",
        "components": ["premium", "fraud", "forecast"]
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
