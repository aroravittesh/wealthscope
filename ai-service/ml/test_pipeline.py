"""Training + inference smoke tests (run from repo: python test_pipeline.py)."""
from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

import joblib

ML_ROOT = Path(__file__).resolve().parent


class TestTrainAndPredict(unittest.TestCase):
    def test_train_to_temp_and_predict(self):
        from train_intent_model import train_and_save

        with tempfile.TemporaryDirectory() as td:
            out = Path(td)
            train_and_save(ML_ROOT / "intent_dataset.csv", out, verbose=False)
            self.assertTrue((out / "intent_model.joblib").is_file())
            vec = joblib.load(out / "intent_vectorizer.joblib")
            model = joblib.load(out / "intent_model.joblib")
            x = vec.transform(["What is the stock price of Apple?"])
            label = model.predict(x)[0]
            self.assertIn(label, model.classes_)


class TestFlaskContract(unittest.TestCase):
    def test_classify_intent_endpoint(self):
        if not (ML_ROOT / "intent_model.joblib").is_file():
            self.skipTest("bundled model missing; run train_intent_model.py first")

        import intent_inference_service as inf

        app = inf.app
        client = app.test_client()
        r = client.post("/classify-intent", json={"message": "What is the stock price of AAPL?"})
        self.assertEqual(r.status_code, 200, r.get_data(as_text=True))
        body = r.get_json()
        self.assertIn("intent", body)
        self.assertIn("confidence", body)
        self.assertIn(body["intent"], inf.ALLOWED_LABELS)
        self.assertGreaterEqual(body["confidence"], 0.0)
        self.assertLessEqual(body["confidence"], 1.0)

    def test_empty_message_400(self):
        if not (ML_ROOT / "intent_model.joblib").is_file():
            self.skipTest("bundled model missing")

        from intent_inference_service import app

        r = app.test_client().post("/classify-intent", json={"message": "   "})
        self.assertEqual(r.status_code, 400)


if __name__ == "__main__":
    unittest.main()
