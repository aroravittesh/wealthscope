import pandas as pd
import random

SYNONYMS = {
    "price": ["value", "cost"],
    "stock": ["share"],
    "market": ["trading"],
    "risk": ["danger"],
    "news": ["updates"],
}

def augment_text(text):
    words = text.split()
    new_words = words.copy()

    for i, w in enumerate(words):
        if w.lower() in SYNONYMS and random.random() < 0.3:
            new_words[i] = random.choice(SYNONYMS[w.lower()])

    if len(new_words) > 3 and random.random() < 0.3:
        i = random.randint(0, len(new_words)-2)
        new_words[i], new_words[i+1] = new_words[i+1], new_words[i]

    return " ".join(new_words)


df = pd.read_csv("intent_dataset.csv")

augmented = []

for _, row in df.iterrows():
    text, label = row["text"], row["label"]

    augmented.append((text, label))

    for _ in range(2):
        augmented.append((augment_text(text), label))

aug_df = pd.DataFrame(augmented, columns=["text", "label"])
aug_df.to_csv("augmented_dataset.csv", index=False)

print("✅ Done. New size:", len(aug_df))
