package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"uuid"
)

const fixtureRecordCount = 4

//go:embed data.json
var fixtureJSON []byte

type fixture struct {
	Users                  []userFixture                   `json:"users"`
	Sessions               []sessionFixture                `json:"sessions"`
	LoginThrottles         []loginThrottleFixture          `json:"login_throttles"`
	Categories             []categoryFixture               `json:"categories"`
	Products               []productFixture                `json:"products"`
	SKUs                   []skuFixture                    `json:"skus"`
	Inventories            []inventoryFixture              `json:"inventories"`
	InventoryAdjustments   []inventoryAdjustmentFixture    `json:"inventory_adjustments"`
	FlashSaleCampaigns     []flashSaleCampaignFixture      `json:"flash_sale_campaigns"`
	FlashSaleItems         []flashSaleItemFixture          `json:"flash_sale_items"`
	FlashSaleCustomerUsage []flashSaleCustomerUsageFixture `json:"flash_sale_customer_usage"`
	Carts                  []cartFixture                   `json:"carts"`
	CartItems              []cartItemFixture               `json:"cart_items"`
}

type userFixture struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type sessionFixture struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	TokenSeed string `json:"token_seed"`
	CSRFSeed  string `json:"csrf_seed"`
}

type loginThrottleFixture struct {
	EmailNormalized string `json:"email_normalized"`
	FailureCount    int    `json:"failure_count"`
}

type categoryFixture struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	IsVisible bool   `json:"is_visible"`
}

type productFixture struct {
	ID          string `json:"id"`
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Status      string `json:"status"`
}

type skuFixture struct {
	ID                  string          `json:"id"`
	ProductID           string          `json:"product_id"`
	Code                string          `json:"code"`
	Attributes          json.RawMessage `json:"attributes"`
	RegularPrice        int64           `json:"regular_price"`
	MaxQuantityPerOrder int             `json:"max_quantity_per_order"`
	Status              string          `json:"status"`
}

type inventoryFixture struct {
	SKUID    string `json:"sku_id"`
	OnHand   int64  `json:"on_hand"`
	Reserved int64  `json:"reserved"`
	Version  int64  `json:"version"`
}

type inventoryAdjustmentFixture struct {
	ID           string `json:"id"`
	SKUID        string `json:"sku_id"`
	ActorUserID  string `json:"actor_user_id"`
	Delta        int64  `json:"delta"`
	OnHandBefore int64  `json:"on_hand_before"`
	OnHandAfter  int64  `json:"on_hand_after"`
	Reason       string `json:"reason"`
}

type flashSaleCampaignFixture struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type flashSaleItemFixture struct {
	ID            string `json:"id"`
	CampaignID    string `json:"campaign_id"`
	SKUID         string `json:"sku_id"`
	SalePrice     int64  `json:"sale_price"`
	QuotaTotal    int64  `json:"quota_total"`
	QuotaReserved int64  `json:"quota_reserved"`
	QuotaSold     int64  `json:"quota_sold"`
	PurchaseLimit int    `json:"purchase_limit"`
}

type flashSaleCustomerUsageFixture struct {
	FlashSaleItemID  string `json:"flash_sale_item_id"`
	UserID           string `json:"user_id"`
	ReservedQuantity int64  `json:"reserved_quantity"`
	SoldQuantity     int64  `json:"sold_quantity"`
}

type cartFixture struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

type cartItemFixture struct {
	ID                  string  `json:"id"`
	CartID              string  `json:"cart_id"`
	SKUID               string  `json:"sku_id"`
	Quantity            int     `json:"quantity"`
	IntendedPriceSource string  `json:"intended_price_source"`
	FlashSaleItemID     *string `json:"flash_sale_item_id"`
	Version             int64   `json:"version"`
}

func loadFixture() (fixture, error) {
	var data fixture
	if err := json.Unmarshal(fixtureJSON, &data); err != nil {
		return fixture{}, fmt.Errorf("decode embedded fixture: %w", err)
	}
	return data, nil
}

func validateFixture(data fixture) error {
	if err := validateFixtureCount("users", len(data.Users), 2); err != nil {
		return err
	}
	for name, count := range map[string]int{
		"sessions":                  len(data.Sessions),
		"login_throttles":           len(data.LoginThrottles),
		"categories":                len(data.Categories),
		"products":                  len(data.Products),
		"skus":                      len(data.SKUs),
		"inventories":               len(data.Inventories),
		"inventory_adjustments":     len(data.InventoryAdjustments),
		"flash_sale_campaigns":      len(data.FlashSaleCampaigns),
		"flash_sale_items":          len(data.FlashSaleItems),
		"flash_sale_customer_usage": len(data.FlashSaleCustomerUsage),
		"carts":                     len(data.Carts),
		"cart_items":                len(data.CartItems),
	} {
		if err := validateFixtureCount(name, count, fixtureRecordCount); err != nil {
			return err
		}
	}

	users := make(map[string]userFixture, len(data.Users))
	for _, user := range data.Users {
		if err := requireUUID("user", user.ID); err != nil {
			return err
		}
		if _, exists := users[user.ID]; exists {
			return fmt.Errorf("duplicate user id %q", user.ID)
		}
		if user.Email == "" || user.Password == "" || user.Status != "active" {
			return fmt.Errorf("invalid user %q", user.ID)
		}
		if user.Role != "admin" && user.Role != "customer" {
			return fmt.Errorf("invalid role for user %q", user.ID)
		}
		users[user.ID] = user
	}
	if user, ok := usersByEmail(users, "admin@gmail.com"); !ok || user.Role != "admin" || user.Password != "cloudian" {
		return fmt.Errorf("missing required admin fixture")
	}
	if user, ok := usersByEmail(users, "customer@gmail.com"); !ok || user.Role != "customer" || user.Password != "cloudian" {
		return fmt.Errorf("missing required customer fixture")
	}

	categories := make(map[string]struct{}, len(data.Categories))
	for _, category := range data.Categories {
		if err := addUUID(categories, "category", category.ID); err != nil {
			return err
		}
		if strings.TrimSpace(category.Name) == "" || strings.TrimSpace(category.Slug) != category.Slug {
			return fmt.Errorf("invalid category %q", category.ID)
		}
	}
	products := make(map[string]struct{}, len(data.Products))
	for _, product := range data.Products {
		if err := addUUID(products, "product", product.ID); err != nil {
			return err
		}
		if _, ok := categories[product.CategoryID]; !ok {
			return fmt.Errorf("product %q references missing category %q", product.ID, product.CategoryID)
		}
		if product.Status != "active" && product.Status != "inactive" {
			return fmt.Errorf("invalid product status for %q", product.ID)
		}
	}
	skus := make(map[string]skuFixture, len(data.SKUs))
	for _, sku := range data.SKUs {
		if err := requireUUID("sku", sku.ID); err != nil {
			return err
		}
		if _, exists := skus[sku.ID]; exists {
			return fmt.Errorf("duplicate sku id %q", sku.ID)
		}
		if _, ok := products[sku.ProductID]; !ok {
			return fmt.Errorf("sku %q references missing product %q", sku.ID, sku.ProductID)
		}
		if sku.RegularPrice <= 0 || sku.MaxQuantityPerOrder <= 0 || !jsonObject(sku.Attributes) {
			return fmt.Errorf("invalid sku %q", sku.ID)
		}
		skus[sku.ID] = sku
	}

	inventories := make(map[string]inventoryFixture, len(data.Inventories))
	for _, inventory := range data.Inventories {
		if _, ok := skus[inventory.SKUID]; !ok {
			return fmt.Errorf("inventory references missing sku %q", inventory.SKUID)
		}
		if _, exists := inventories[inventory.SKUID]; exists || inventory.Reserved < 0 || inventory.Reserved > inventory.OnHand || inventory.Version < 0 {
			return fmt.Errorf("invalid inventory for sku %q", inventory.SKUID)
		}
		inventories[inventory.SKUID] = inventory
	}
	for _, adjustment := range data.InventoryAdjustments {
		if _, ok := inventories[adjustment.SKUID]; !ok {
			return fmt.Errorf("adjustment %q references missing inventory", adjustment.ID)
		}
		if _, ok := users[adjustment.ActorUserID]; !ok {
			return fmt.Errorf("adjustment %q references missing actor", adjustment.ID)
		}
		if adjustment.Delta == 0 || adjustment.OnHandAfter != adjustment.OnHandBefore+adjustment.Delta || strings.TrimSpace(adjustment.Reason) == "" {
			return fmt.Errorf("invalid adjustment %q", adjustment.ID)
		}
		if inventory := inventories[adjustment.SKUID]; inventory.OnHand != adjustment.OnHandAfter {
			return fmt.Errorf("adjustment %q does not match inventory", adjustment.ID)
		}
	}

	campaigns := make(map[string]struct{}, len(data.FlashSaleCampaigns))
	for _, campaign := range data.FlashSaleCampaigns {
		if err := addUUID(campaigns, "flash sale campaign", campaign.ID); err != nil {
			return err
		}
		if strings.TrimSpace(campaign.Name) == "" || campaign.Status != "draft" {
			return fmt.Errorf("invalid flash sale campaign %q", campaign.ID)
		}
	}
	items := make(map[string]flashSaleItemFixture, len(data.FlashSaleItems))
	for _, item := range data.FlashSaleItems {
		if err := requireUUID("flash sale item", item.ID); err != nil {
			return err
		}
		if _, ok := campaigns[item.CampaignID]; !ok {
			return fmt.Errorf("flash sale item %q references missing campaign", item.ID)
		}
		sku, ok := skus[item.SKUID]
		if !ok || item.SalePrice <= 0 || item.SalePrice >= sku.RegularPrice || item.QuotaTotal <= 0 || item.QuotaReserved < 0 || item.QuotaSold < 0 || item.QuotaReserved+item.QuotaSold > item.QuotaTotal || item.PurchaseLimit <= 0 {
			return fmt.Errorf("invalid flash sale item %q", item.ID)
		}
		items[item.ID] = item
	}
	for _, usage := range data.FlashSaleCustomerUsage {
		item, ok := items[usage.FlashSaleItemID]
		if !ok {
			return fmt.Errorf("usage references missing flash sale item %q", usage.FlashSaleItemID)
		}
		if _, ok := users[usage.UserID]; !ok || usage.ReservedQuantity < 0 || usage.SoldQuantity < 0 || usage.ReservedQuantity+usage.SoldQuantity > int64(item.PurchaseLimit) {
			return fmt.Errorf("invalid flash sale usage for %q", usage.FlashSaleItemID)
		}
	}

	carts := make(map[string]cartFixture, len(data.Carts))
	activeCarts := make(map[string]int)
	for _, cart := range data.Carts {
		if _, ok := users[cart.UserID]; !ok || (cart.Status != "active" && cart.Status != "archived") {
			return fmt.Errorf("invalid cart %q", cart.ID)
		}
		if _, exists := carts[cart.ID]; exists {
			return fmt.Errorf("duplicate cart id %q", cart.ID)
		}
		if cart.Status == "active" {
			activeCarts[cart.UserID]++
		}
		carts[cart.ID] = cart
	}
	customer, _ := usersByEmail(users, "customer@gmail.com")
	if activeCarts[customer.ID] != 1 {
		return fmt.Errorf("fixture must have one active customer cart")
	}
	for _, item := range data.CartItems {
		if _, ok := carts[item.CartID]; !ok {
			return fmt.Errorf("cart item %q references missing cart", item.ID)
		}
		if _, ok := skus[item.SKUID]; !ok || item.Quantity <= 0 || item.Version < 1 {
			return fmt.Errorf("invalid cart item %q", item.ID)
		}
		if item.IntendedPriceSource == "regular" && item.FlashSaleItemID != nil {
			return fmt.Errorf("regular cart item %q has flash sale item", item.ID)
		}
		if item.IntendedPriceSource == "flash_sale" {
			if item.FlashSaleItemID == nil || items[*item.FlashSaleItemID].SKUID != item.SKUID {
				return fmt.Errorf("invalid flash sale cart item %q", item.ID)
			}
		}
	}

	return nil
}

func validateFixtureCount(name string, actual, expected int) error {
	if actual != expected {
		return fmt.Errorf("%s count = %d, want %d", name, actual, expected)
	}
	return nil
}

func usersByEmail(users map[string]userFixture, email string) (userFixture, bool) {
	for _, user := range users {
		if user.Email == email {
			return user, true
		}
	}
	return userFixture{}, false
}

func addUUID(values map[string]struct{}, label, value string) error {
	if err := requireUUID(label, value); err != nil {
		return err
	}
	if _, exists := values[value]; exists {
		return fmt.Errorf("duplicate %s id %q", label, value)
	}
	values[value] = struct{}{}
	return nil
}

func requireUUID(label, value string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("invalid %s id %q", label, value)
	}
	return nil
}

func jsonObject(value json.RawMessage) bool {
	var decoded map[string]any
	return json.Unmarshal(value, &decoded) == nil && decoded != nil
}
