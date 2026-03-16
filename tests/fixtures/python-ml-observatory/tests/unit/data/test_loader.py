"""Tests for data loader."""
from src.data.loader import load_training_data, load_eval_dataset, split_dataset

def test_load_training_data():
    data = load_training_data("dummy.csv")
    assert len(data) > 0

def test_load_eval_dataset():
    data = load_eval_dataset()
    assert len(data) == 3

def test_split_dataset():
    data = [1, 2, 3, 4, 5]
    train, test = split_dataset(data, 0.6)
    assert len(train) == 3
    assert len(test) == 2
