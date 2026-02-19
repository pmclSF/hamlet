import pytest

from app.services.inventory import InventoryService, OutOfStockError


@pytest.fixture
def warehouse():
    service = InventoryService()
    service.add_stock('WIDGET-A', 100)
    service.add_stock('WIDGET-B', 5)
    return service


def test_check_stock_returns_quantity(warehouse):
    assert warehouse.check_stock('WIDGET-A') == 100


def test_reserve_reduces_available_stock(warehouse):
    warehouse.reserve('WIDGET-A', 10)
    assert warehouse.check_stock('WIDGET-A') == 90


def test_reserve_raises_when_insufficient(warehouse):
    with pytest.raises(OutOfStockError, match='Insufficient stock for WIDGET-B'):
        warehouse.reserve('WIDGET-B', 20)


@pytest.mark.parametrize('sku, initial, reserve, expected', [
    ('WIDGET-A', 100, 1, 99),
    ('WIDGET-A', 100, 50, 50),
    ('WIDGET-B', 5, 5, 0),
])
def test_reserve_various_quantities(warehouse, sku, initial, reserve, expected):
    warehouse.reserve(sku, reserve)
    assert warehouse.check_stock(sku) == expected


def test_restock_increases_quantity(warehouse):
    warehouse.restock('WIDGET-B', 15)
    assert warehouse.check_stock('WIDGET-B') == 20
