//Main GoRecurly Package
package gorecurly

//TODO: Do all tests
//TODO: Check all comments when finished 
//TODO: Check that state is working with lists
//TODO: Introduce stubs for all resources
//TODO: Invoice resources
//TODO: PDF Invoice
//TODO: Plans resources
//TODO: Add ons resources
//TODO: Subscriptions resources
//TODO: Transactions resources
//TODO: Recurly.js signing
//TODO: commit patch to properly encode dates

import (
	"net/http"
	"io"
	"errors"
	"bytes"
	"io/ioutil"
	"fmt"
	"time"
	"strings"
	"encoding/xml"
	"net/url"
)

const (
	URL = "https://api.recurly.com/v2/"
	libversion = "0.1"
	libname = "Recurly-Go"
	ACCOUNTS = "accounts"
	ADJUSTMENTS = "adjustments"
	BILLINGINFO = "billing_info"
	COUPONS = "coupons"
	COUPONREDEMPTIONS = "redemption"
	INVOICES = "invoices"
	PLANS = "plans"
	PLANADDONS = "add_ons"
	SUBSCRIPTIONS = "subscriptions"
	TRANSACTIONS = "transactions"

)
//Generic Reader
type nopCloser struct {
	io.Reader
}
//functions

//Initialize the Recurly package with your apikey and your jskey
func InitRecurly(apikey string,jskey string) (*Recurly){
	r := new (Recurly)
	r.apiKey = apikey
	r.JSKey = jskey
	return r
}

//interfaces


//Paging interface to allow Next,Prev,Start
type Pager interface {
	getRawBody() []byte
}

//recurly errors
var Error400 = errors.New("The request was invalid or could not be understood by the server. Resubmitting the request will likely result in the same error.")
var Error401 = errors.New("Your API key is missing or invalid.")
var Error402 = errors.New("Your Recurly account is in production mode but is not in good standing. Please pay any outstanding invoices.")
var Error403 = errors.New("The login is attempting to perform an action it does not have privileges to access. Verify your login credentials are for the appropriate account.")
var Error404 = errors.New("The resource was not found with the given identifier. The response body will explain which resource was not found.")
var Error405 = errors.New("The requested method is not valid at the given URL.")
var Error406 = errors.New("The request's Accept header is not set to application/xml")
var Error412 = errors.New("The request was unsuccessful because a condition was not met. For example, this message may be returned if you attempt to cancel a subscription for an account that has no subscription.")
var Error429 = errors.New("You have made too many API requests in the last hour. Future API requests will be ignored until the beginning of the next hour.")

//Recurly Generic Errors
type RecurlyError struct {
	XMLName xml.Name `xml:"error"`
	statusCode int
	Symbol string `xml:"symbol"`
	Description string `xml:"description"`
	Details string `xml:"details"`
}

//Recurly Validation Errors Array
type RecurlyValidationErrors struct {
	XMLName xml.Name `xml:"errors"`
	statusCode int
	Errors []RecurlyValidationError `xml:"error"`
}

//Recurly validation error
type RecurlyValidationError struct {
	XMLName xml.Name `xml:"error"`
	FieldName string `xml:"field,attr"`
	Symbol string `xml:"symbol,attr"`
	Description string `xml:",innerxml"`
}

//Parse Recurly XML to create a Recurly Error
func CreateRecurlyStandardError(resp *http.Response) (r RecurlyError) {
	r.statusCode = resp.StatusCode
	if xmlstring, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
		if xmlerr := xml.Unmarshal(xmlstring, &r); xmlerr != nil {
			r.Description = string(xmlstring)
		}
	}
	return r
}

//Parse Recurly XML to create a Validation Error
func CreateRecurlyValidationError(resp *http.Response) (r RecurlyValidationErrors) {
	r.statusCode = resp.StatusCode
	if xmlstring, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
		println(string(xmlstring))
		if xmlerr := xml.Unmarshal(xmlstring, &r); xmlerr != nil {
			//r.Description = xmlerr.Error()
			println(xmlerr.Error())
		}
	}
	return r
}

//Filter to decide which error type to create
func createRecurlyError(resp *http.Response) ( error) {
	switch resp.StatusCode {
	case 400:
		return Error400
	case 401:
		return Error401
	case 402:
		return Error402
	case 403:
		return Error403
	case 404:
		return Error404
	case 405:
		return Error405
	case 406:
		return Error406
	case 412:
		return Error412
	case 429:
		return Error429
	case 422 :
		return CreateRecurlyStandardError(resp)
	}
	return CreateRecurlyValidationError(resp)
}

//Formatted General Error 
func (r RecurlyError) Error() string {
	return fmt.Sprintf("Recurly Error: %s , %s %s Status Code: %v", r.Symbol,r.Description, r.Details,r.statusCode)
}

//Formatted Validation Error
func (r RecurlyValidationErrors) Error() string {
	var rtnString string
	for _,v := range r.Errors {
		rtnString += v.FieldName + " " + v.Description + "\n"
	}
	return fmt.Sprintf("You have the following validation errors:\n%s", rtnString)
}

//Main Recurly Client
type Recurly struct {
	apiKey, JSKey  string
	debug bool
}

//Set verbose debugging
func (r *Recurly) EnableDebug() {
	r.debug = true
}

//Get a list of accounts
func (r *Recurly) GetAccounts(params ...url.Values) (AccountList, error){
	accountlist := AccountList{}
	sendvars := url.Values{}
	if params != nil {
		sendvars = params[0] 
		accountlist.perPage = sendvars.Get("per_page")
	} 
	if err := accountlist.initList(ACCOUNTS,sendvars,r); err == nil {
		if xmlerr := xml.Unmarshal(accountlist.getRawBody(), &accountlist); xmlerr == nil {
			for k,_ := range accountlist.Account {
				accountlist.Account[k].r = r
				accountlist.Account[k].endpoint = ACCOUNTS
			}
			accountlist.r = r
			return accountlist, nil
		} else {
			if r.debug {
				println(xmlerr.Error())
			}
			return accountlist, xmlerr
		}
	} else {
		return accountlist, err
	}
	return accountlist, nil
}

//Get a list of adjustments for an account
func (r *Recurly) GetAdjustments(account_code string,params ...url.Values) (AdjustmentList, error){
	adjlist := AdjustmentList{}
	sendvars := url.Values{}
	if params != nil {
		sendvars = params[0] 
		adjlist.perPage = sendvars.Get("per_page")
	} 
	if err := adjlist.initList(ACCOUNTS + "/" + account_code + "/"  + ADJUSTMENTS,sendvars,r); err == nil {
		if xmlerr := xml.Unmarshal(adjlist.getRawBody(), &adjlist); xmlerr == nil {
			for k,_ := range adjlist.Adjustments {
				adjlist.Adjustments[k].r = r
				adjlist.Adjustments[k].endpoint = ADJUSTMENTS
			}
			adjlist.r = r
			adjlist.AccountCode = account_code
			return adjlist, nil
		} else {
			if r.debug {
				println(xmlerr.Error())
			}
			return adjlist, xmlerr
		}
	} else {
		return adjlist, err
	}
	return adjlist, nil
}

//Get a list of coupons
func (r *Recurly) GetCoupons(params ...url.Values) (CouponList, error){
	cplist := CouponList{}
	sendvars := url.Values{}
	if params != nil {
		sendvars = params[0] 
		cplist.perPage = sendvars.Get("per_page")
	} 
	if err := cplist.initList(COUPONS,sendvars,r); err == nil {
		if xmlerr := xml.Unmarshal(cplist.getRawBody(), &cplist); xmlerr == nil {
			for k,_ := range cplist.Coupons {
				cplist.Coupons[k].r = r
				cplist.Coupons[k].endpoint = COUPONS
			}
			cplist.r = r
			return cplist, nil
		} else {
			if r.debug {
				println(xmlerr.Error())
			}
			return cplist, xmlerr
		}
	} else {
		return cplist, err
	}
	return cplist, nil
}

//Get a list of invoices for an account
func (r *Recurly) GetAccountInvoices(account_code string, params ...url.Values) (AccountInvoiceList, error){
	invoicelist := AccountInvoiceList{}
	sendvars := url.Values{}
	if params != nil {
		sendvars = params[0] 
		invoicelist.perPage = sendvars.Get("per_page")
	} 
	if err := invoicelist.initList(ACCOUNTS + "/" + account_code + "/" + INVOICES,sendvars,r); err == nil {
		if xmlerr := xml.Unmarshal(invoicelist.getRawBody(), &invoicelist); xmlerr == nil {
			for k,_ := range invoicelist.Invoices {
				invoicelist.Invoices[k].r = r
				invoicelist.Invoices[k].endpoint = INVOICES
			}
			invoicelist.r = r
			invoicelist.AccountCode = account_code
			return invoicelist, nil
		} else {
			if r.debug {
				println(xmlerr.Error())
			}
			return invoicelist, xmlerr
		}
	} else {
		return invoicelist, err
	}
	return invoicelist, nil
}

//Get a list of invoices
func (r *Recurly) GetInvoices(params ...url.Values) (InvoiceList, error){
	invoicelist := InvoiceList{}
	sendvars := url.Values{}
	if params != nil {
		sendvars = params[0] 
		invoicelist.perPage = sendvars.Get("per_page")
	} 
	if err := invoicelist.initList(INVOICES,sendvars,r); err == nil {
		if xmlerr := xml.Unmarshal(invoicelist.getRawBody(), &invoicelist); xmlerr == nil {
			for k,_ := range invoicelist.Invoices {
				invoicelist.Invoices[k].r = r
				invoicelist.Invoices[k].endpoint = INVOICES
			}
			invoicelist.r = r
			return invoicelist, nil
		} else {
			if r.debug {
				println(xmlerr.Error())
			}
			return invoicelist, xmlerr
		}
	} else {
		return invoicelist, err
	}
	return invoicelist, nil
}

//Get a single account by key
func (r *Recurly) GetAccount(account_code string) (account Account, err error) {
	account = r.NewAccount()
	if resp,err := r.createRequest(ACCOUNTS + "/" + account_code,"GET", nil, nil); err == nil {
		if resp.StatusCode == 200 {
			if body, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
				if r.debug {
					println(resp.Status)	
					for k, _ := range resp.Header {
						println(k + ":" + resp.Header[k][0])
					}
					fmt.Printf("%s\n", body) 
					fmt.Printf("Content-Length:%v\n", resp.ContentLength) 
				}
				//load object xml
				if xmlerr := xml.Unmarshal(body, &account); xmlerr != nil {
					return account,xmlerr
				}
				//everything went fine
				return  account,nil
			} else {
				//return read error
				return account,readerr
			}
			return account,nil
		} else {
			return account,createRecurlyError(resp)
		}
	} else {
		return account, err
	}
	return account, nil
}

//Get a single account by key
func (r *Recurly) GetAdjustment(uuid string) (adj Adjustment, err error) {
	adj = r.NewAdjustment()
	if resp,err := r.createRequest(ADJUSTMENTS + "/" + uuid,"GET", nil, nil); err == nil {
		if resp.StatusCode == 200 {
			if body, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
				if r.debug {
					println(resp.Status)	
					for k, _ := range resp.Header {
						println(k + ":" + resp.Header[k][0])
					}
					fmt.Printf("%s\n", body) 
					fmt.Printf("Content-Length:%v\n", resp.ContentLength) 
				}
				//load object xml
				if xmlerr := xml.Unmarshal(body, &adj); xmlerr != nil {
					return adj,xmlerr
				}
				//everything went fine
				return  adj,nil
			} else {
				//return read error
				return adj,readerr
			}
			return adj,nil
		} else {
			return adj,createRecurlyError(resp)
		}
	} else {
		return adj, err
	}
	return adj, nil
}
//Get a single account by key
func (r *Recurly) GetCouponRedemption(account_code string) (red Redemption, err error) {
	red.r = r
	red.AccountCode = account_code
	if resp,err := r.createRequest(ACCOUNTS + "/" + account_code + "/redemption","GET", nil, nil); err == nil {
		if resp.StatusCode == 200 {
			if body, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
				if r.debug {
					println(resp.Status)	
					for k, _ := range resp.Header {
						println(k + ":" + resp.Header[k][0])
					}
					fmt.Printf("%s\n", body) 
					fmt.Printf("Content-Length:%v\n", resp.ContentLength) 
				}
				//load object xml
				if xmlerr := xml.Unmarshal(body, &red); xmlerr != nil {
					return red,xmlerr
				}
				//everything went fine
				return  red,nil
			} else {
				//return read error
				return red,readerr
			}
			return red,nil
		} else {
			return red,createRecurlyError(resp)
		}
	} else {
		return red, err
	}
	return red, nil
}
//Get a single coupon
func (r *Recurly) GetCoupon(uuid string) (coupon Coupon, err error) {
	coupon = r.NewCoupon()
	if resp,err := r.createRequest(COUPONS + "/" + uuid,"GET", nil, nil); err == nil {
		if resp.StatusCode == 200 {
			if body, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
				if r.debug {
					println(resp.Status)	
					for k, _ := range resp.Header {
						println(k + ":" + resp.Header[k][0])
					}
					fmt.Printf("%s\n", body) 
					fmt.Printf("Content-Length:%v\n", resp.ContentLength) 
				}
				//load object xml
				if xmlerr := xml.Unmarshal(body, &coupon); xmlerr != nil {
					return coupon,xmlerr
				}
				//everything went fine
				return  coupon,nil
			} else {
				//return read error
				return coupon,readerr
			}
			return coupon,nil
		} else {
			return coupon,createRecurlyError(resp)
		}
	} else {
		return coupon, err
	}
	return coupon, nil
}

//Get invoice by uuid
func (r *Recurly) GetInvoice(uuid string) (invoice Invoice, err error) {
	invoice = r.NewInvoice()
	if resp,err := r.createRequest(INVOICES + "/" + uuid,"GET", nil, nil); err == nil {
		if resp.StatusCode == 200 {
			if body, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
				if r.debug {
					println(resp.Status)	
					for k, _ := range resp.Header {
						println(k + ":" + resp.Header[k][0])
					}
					fmt.Printf("%s\n", body) 
					fmt.Printf("Content-Length:%v\n", resp.ContentLength) 
				}
				//load object xml
				if xmlerr := xml.Unmarshal(body, &invoice); xmlerr != nil {
					return invoice,xmlerr
				}
				//everything went fine
				return  invoice,nil
			} else {
				//return read error
				return invoice,readerr
			}
			return invoice,nil
		} else {
			return invoice,createRecurlyError(resp)
		}
	} else {
		return invoice, err
	}
	return invoice, nil
}

//Create a new Account
func (r *Recurly) NewAccount() (account Account) {
	account.r = r
	account.endpoint = ACCOUNTS
	return
}

//Create a new Adjustment
func (r *Recurly) NewAdjustment() (adj Adjustment) {
	adj.r = r
	adj.endpoint = ADJUSTMENTS
	return
}

//Create new Billing Info
func (r *Recurly) NewBillingInfo() (bi BillingInfo) {
	bi.r = r
	bi.endpoint = BILLINGINFO
	return
}

//Create a new Coupon
func (r *Recurly) NewCoupon() (c Coupon) {
	c.r = r
	c.endpoint = COUPONS
	return
}

//Invoice Pending Charges on an account
func (r *Recurly) InvoicePendingCharges(account_code string) (invoice Invoice, e error) {
	invoice.r = r
	e = invoice.r.doCreate(&invoice,ACCOUNTS + "/" + account_code + "/invoices")
	return
}

//Create a new Invoice
func (r *Recurly) NewInvoice() (invoice Invoice) {
	invoice.r = r
	invoice.endpoint = INVOICES
	return
}

//Get a single accounts billing info by key
func (r *Recurly) GetBillingInfo(account_code string) (bi BillingInfo, err error) {
	bi = r.NewBillingInfo()
	if resp,err := r.createRequest(ACCOUNTS + "/" + account_code + "/" + BILLINGINFO,"GET", nil, nil); err == nil {
		if resp.StatusCode == 200 {
			if body, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
				if r.debug {
					println(resp.Status)	
					for k, _ := range resp.Header {
						println(k + ":" + resp.Header[k][0])
					}
					fmt.Printf("%s\n", body) 
					fmt.Printf("Content-Length:%v\n", resp.ContentLength) 
				}
				//load object xml
				if xmlerr := xml.Unmarshal(body, &bi); xmlerr != nil {
					return bi,xmlerr
				}
				//everything went fine
				bi.Account.endpoint = ACCOUNTS
				return  bi,nil
			} else {
				//return read error
				return bi,readerr
			}
			return bi,nil
		} else {
			return bi,createRecurlyError(resp)
		}
	} else {
		return bi, err
	}
	return bi, nil
}
//Create a request to Recurly and return that response object
func (r *Recurly) createRequest(endpoint string, method string, params url.Values, msgbody []byte) (*http.Response, error) { 
	client := &http.Client{}

	u, err := url.Parse(URL + endpoint)
	if err != nil {
		return nil,err
	}
	u.RawQuery = params.Encode()
	body := nopCloser{bytes.NewBufferString(string(msgbody))}
	if r.debug {
		fmt.Printf("Endpoint Requested: %s Method: %s Body: %s\n", u.String(), method, string(msgbody))
	}
	if req, err := http.NewRequest(method, u.String(), body); err != nil {
		return nil,err
	} else {
		req.Header.Add("Accept", "application/xml")
		req.Header.Add("Accept-Language", "en-US")
		req.Header.Add("User-Agent", libname + " version=" + libversion)
		req.Header.Add("Content-Type","application/xml; charset=utf-8")
		req.ContentLength = int64(len(string(msgbody)))
		req.SetBasicAuth(r.apiKey,"")
		if resp, resperr := client.Do(req); resperr == nil {
			return resp, nil
		} else {
			return nil,resperr
		}
	}
	return nil, nil
}

//Create a resource from struct, uses POST method
func (r *Recurly) doCreate(v interface{}, endpoint string) (error) {
	if xmlstring, err := xml.MarshalIndent(v, "", "    "); err == nil {
		xmlstring = []byte(xml.Header + string(xmlstring))
		if r.debug {
			fmt.Printf("%s\n",xmlstring)
		}
		if resp, reqerr := r.createRequest(endpoint, "POST", nil, xmlstring); reqerr == nil {
			if resp.StatusCode < 400 {
				if body, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
					if r.debug {
						println(resp.Status)	
						for k, _ := range resp.Header {
							println(k + ":" + resp.Header[k][0])
						}
						fmt.Printf("%s\n", body) 
						fmt.Printf("Content-Length:%v\n", resp.ContentLength) 
					}
					//load object xml
					if xmlerr := xml.Unmarshal(body, v); xmlerr != nil {
						return xmlerr
					}
					//everything went fine
					return  nil
				} else {
					//return read error
					return readerr
				}
				return nil
			} else {
				return createRecurlyError(resp)
			}
		} else {
			return reqerr
		}
	} else {
		return err
	}
	return nil
}

//Update a resource from Struct, uses PUT method
func (r *Recurly) doUpdate(v interface{}, endpoint string) (error) {
	if xmlstring, err := xml.MarshalIndent(v, "", "    "); err == nil {
		xmlstring = []byte(xml.Header + string(xmlstring))
		if r.debug {
			fmt.Printf("%s\n",xmlstring)
		}
		if resp, reqerr := r.createRequest(endpoint, "PUT", nil, xmlstring); reqerr == nil {
			if resp.StatusCode < 400 {
				return nil
			} else {
				return createRecurlyError(resp)
			}
		} else {
			return reqerr
		}
	} else {
		return err
	}
	return nil
}

//Delete a resource, uses DELETE method
func (r *Recurly) doDelete(endpoint string) (error) {
	if resp, reqerr := r.createRequest(endpoint, "DELETE", nil, nil); reqerr == nil {
		if resp.StatusCode < 400 {
			return nil
		} else {
			return createRecurlyError(resp)
		}
	} else {
		return reqerr
	}
	return nil
}

/* paging struct */

//A struct to assist in paging result sets
type Paging struct {
	rawBody []byte
	count, next, prev, perPage string
}

//Return the rawBody Var
func (p Paging) getRawBody() ([]byte) {
	return p.rawBody
}

//Set header data for paging
func (p *Paging) SetData(rb []byte, count string, links string) {
	p.rawBody = rb
	p.count = count
	p.next = ""
	p.prev = ""
	for _,v := range strings.SplitN(links,",",-1) {
		println(v)
		link := strings.SplitN(v,";",-1)
		link[0] = strings.Replace(link[0],"<","",-1)
		link[0] = strings.Replace(link[0],">","",-1)
		if u, err := url.Parse(link[0]); err == nil {
			values := u.Query() 
			switch link[1] {
			case " rel=\"next\"" :
				p.next = values.Get("cursor")
			case " rel=\"prev\"" :
				p.prev = values.Get("cursor")
			}
		} 
	}
}

//Initialize the paging list values
func (p *Paging) initList(endpoint string, params url.Values, r *Recurly) ( error) { 
	if resp, err := r.createRequest(endpoint,"GET",params, make([]byte,0)); err == nil {
		if resp.StatusCode < 400 {
			defer resp.Body.Close()
			if body, readerr := ioutil.ReadAll(resp.Body); readerr == nil {
				if r.debug {
					println(resp.Status)	
					for k, _ := range resp.Header {
						println(k + ":" + resp.Header[k][0])
					}
					fmt.Printf("%s\n", body) 
					fmt.Printf("Content-Length:%v\n", resp.ContentLength) 
				}
				if x := len(resp.Header["Link"]); x > 0{
					p.SetData(body,resp.Header["X-Records"][0],resp.Header["Link"][0])
				} else {
					p.SetData(body,resp.Header["X-Records"][0],"")
				}
				//everything went fine
				return  nil
			} else {
				//return read error
				return readerr
			}
		} else {
			return createRecurlyError(resp) 
		}
	} else {
		//return error message
		return err
	}
	return nil
}

/*resource objects */

//Account pager
type AccountList struct {
	Paging
	r *Recurly
	XMLName xml.Name `xml:"accounts"`
	Account []Account `xml:"account"`
}

//Get next set of accounts
func (a *AccountList) Next() (bool) {
	if a.next != "" {
		v := url.Values{}
		v.Set("cursor",a.next)
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAccounts(v)
	} else {
		return false
	}
	return true
}

//Get previous set of accounts
func (a *AccountList) Prev() ( bool) {
	if a.prev != "" {
		v := url.Values{}
		v.Set("cursor",a.prev)
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAccounts(v)
	} else {
		return false
	}
	return true
}

//Go to start set of accounts
func (a *AccountList) Start() ( bool) {
	if a.prev != "" {
		v := url.Values{}
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAccounts(v)
	} else {
		return false
	}
	return true
}

//Account struct
type Account struct{
	XMLName xml.Name `xml:"account"`
	endpoint string
	r *Recurly
	AccountCode string `xml:"account_code"`
	Username string `xml:"username"`
	Email string `xml:"email"`
	State string `xml:"state,omitempty"`
	FirstName string `xml:"first_name"`
	LastName string `xml:"last_name"`
	CompanyName string `xml:"company_name"`
	AcceptLanguage string `xml:"accept_language"`
	HostedLoginToken string `xml:"hosted_login_token,omitempty"`
	CreatedAt *time.Time `xml:"created_at,omitempty"`
	B *BillingInfo `xml:"billing_info,omitempty"` 
}

//Create a new account and load updated fields
func (a *Account) Create() (error) {
	if a.CreatedAt != nil || a.HostedLoginToken != "" || a.State != "" {
		return RecurlyError{statusCode:400,Description:"Account Code Already in Use"}
	}
	err := a.r.doCreate(&a,a.endpoint)
	if err == nil {
		a.B = nil
	}
	return err
}

//Update an account 
func (a *Account) Update() (error) {
	newaccount := new(Account)
	*newaccount = *a
	newaccount.State = ""
	newaccount.HostedLoginToken = ""
	newaccount.CreatedAt = nil
	newaccount.B = nil
	return a.r.doUpdate(newaccount,a.endpoint + "/" + a.AccountCode)
}

//Close an account
func (a *Account) Close() (error) {
	return a.r.doDelete(a.endpoint + "/" + a.AccountCode)
}

//Reopen a closed account
func (a *Account) Reopen() (error) {
	newaccount := new(Account)
	return a.r.doUpdate(newaccount,a.endpoint + "/" + a.AccountCode + "/reopen")
}
/* Stub */
type stub struct {
	HREF string `xml:"href,attr"`
	endpoint string `xml:",-"`
}

func (s stub) GetCode() (code string) {
	code = "invalidcode"
	if s.HREF != "" {
		code = strings.Replace(s.HREF,URL + s.endpoint + "/","",-1)
	}
	return 
}

//Account Stub struct
type AccountStub struct {
	XMLName xml.Name `xml:"account"`
	stub
}

//adjustment struct
type Adjustment struct{
	XMLName xml.Name `xml:"adjustment"`
	endpoint string
	r *Recurly
	Type string `xml:"type,attr"`
	AccountCode string `xml:"-"`
	UUID string `xml:"uuid,omitempty"`
	Description string `xml:"description,omitempty"`
	AccountingCode string `xml:"accounting_code,omitempty"`
	Origin string `xml:"origin,omitempty"`
	UnitAmountInCents int `xml:"unit_amount_in_cents,omitempty"`
	Quantity int `xml:"quantity,omitempty"`
	DiscountInCents int `xml:"discount_in_cents,omitempty"`
	TaxInCents int `xml:"tax_in_cents,omitempty"`
	Currency string `xml:"currency,omitempty"`
	Taxable bool `xml:"taxable,omitempty"`
	StartDate *time.Time `xml:"start_date,omitempty"`
	EndDate *time.Time `xml:"end_date,omitempty"`
	CreatedAt *time.Time `xml:"created_at,omitempty"`
}

//Create a new adjustment and load updated fields
func (a *Adjustment) Create() (error) {
	if a.UUID != "" {
		return RecurlyError{statusCode:400,Description:"Adjustment Already created"}
	}
	return a.r.doCreate(&a,ACCOUNTS + "/" + a.AccountCode + "/" + a.endpoint)
}

//delete and adjustment
func (a *Adjustment) Delete() (error) {
	return a.r.doDelete(a.endpoint + "/" + a.UUID)
}
type AdjustmentList struct {
	Paging
	r *Recurly
	AccountCode string
	XMLName xml.Name `xml:"adjustments"`
	Adjustments []Adjustment `xml:"adjustment"`
}

//Get next set of adjustments
func (a *AdjustmentList) Next() (bool) {
	if a.next != "" {
		v := url.Values{}
		v.Set("cursor",a.next)
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAdjustments(a.AccountCode,v)
	} else {
		return false
	}
	return true
}

//Get previous set of accounts
func (a *AdjustmentList) Prev() ( bool) {
	if a.prev != "" {
		v := url.Values{}
		v.Set("cursor",a.prev)
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAdjustments(a.AccountCode,v)
	} else {
		return false
	}
	return true
}

//Go to start set of accounts
func (a *AdjustmentList) Start() ( bool) {
	if a.prev != "" {
		v := url.Values{}
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAdjustments(a.AccountCode,v)
	} else {
		return false
	}
	return true
}

//Billing Info struct
type BillingInfo struct {
	XMLName xml.Name `xml:"billing_info"`
	endpoint string
	r *Recurly
	Account *AccountStub `xml:"account,omitempty"`
	AccountCode string `xml:"account_code,omitempty"`
	FirstName string `xml:"first_name,omitempty"`
	LastName string `xml:"last_name,omitempty"`
	Address1 string `xml:"address1,omitempty"`
	Address2 string `xml:"address2,omitempty"`
	City string `xml:"city,omitempty"`
	State string `xml:"state,omitempty"`
	Zip string `xml:"zip,omitempty"`
	Country string `xml:"country,omitempty"`
	Phone string `xml:"phone,omitempty"`
	VatNumber string `xml:"vat_number,omitempty"`
	IPAddress string `xml:"ip_address,omitempty"`
	IPAddressCountry string `xml:"ip_address_country,omitempty"`
	Number string `xml:"number,omitempty"`
	FirstSix string `xml:"first_six,omitempty"`
	LastFour string `xml:"last_four,omitempty"`
	VerificationValue string `xml:"verification_value,omitempty"`
	CardType string `xml:"card_type,omitempty"`
	Month int `xml:"month,omitempty"`
	Year int `xml:"year,omitempty"`
	BillingAgreementID string `xml:"billing_agreement_id,omitempty"`
}

//Update an billing info 
func (b *BillingInfo) Update() (error) {
	newbilling := new(BillingInfo)
	*newbilling = *b
	newbilling.AccountCode = ""
	newbilling.Account = nil
	newbilling.FirstSix = ""
	newbilling.LastFour = ""
	newbilling.CardType = ""
	return b.r.doUpdate(newbilling,ACCOUNTS + "/" + b.Account.GetCode() + "/" + BILLINGINFO)
}

//Delete billing info for an account
func (b *BillingInfo) Delete() (error) {
	return b.r.doDelete(ACCOUNTS + "/" + b.AccountCode + "/" + BILLINGINFO)
}

func (b BillingInfo) GetAccount() (Account,error) {
	return b.r.GetAccount(b.Account.GetCode())
}

type PlanCode struct {
	XMLName xml.Name `xml:"plan_codes"`
	PlanCode []string `xml:"plan_code"`
}

type Redemption struct {
	XMLName xml.Name `xml:"redemption"`
	r *Recurly
	AccountCode string `xml:"account_code,omitempty"`
	SingleUse bool `xml:"single_use,omitempty"`
	TotalDiscountedInCents int `xml:"total_discounted_in_cents,omitempty"`
	Currency string `xml:"currency,omitempty"`
	CreatedAt *time.Time `xml:"created_at,omitempty"`
}

//delete and redemption
func (r *Redemption) Delete() (error) {
	return r.r.doDelete(ACCOUNTS + "/" + r.AccountCode + "/redemption")
}
type Coupon struct{
	XMLName xml.Name `xml:"coupon"`
	endpoint string
	r *Recurly
	AccountCode string `xml:"-"`
	CouponCode string `xml:"coupon_code"`
	Name string `xml:"name"`
	State string `xml:"state,omitempty"`
	DiscountType string `xml:"discount_type,omitempty"`
	DiscountPercent int `xml:"discount_percent,omitempty"`
	RedeemByDate *time.Time `xml:"redeem_by_date,omitempty"`
	SingleUse bool `xml:"single_use,omitempty"`
	AppliesForMonths string `xml:"applies_for_months,omitempty"`
	MaxRedemptions int `xml:"max_redemptions,omitempty"`
	AppliesToAllPlans bool `xml:"applies_to_all_plans,omitempty"`
	CreatedAt *time.Time `xml:"created_at,omitempty"`
	PlanCodes PlanCode `xml:"plan_codes,omitempty"`
}

//Create a new adjustment and load updated fields
func (c *Coupon) Create() (error) {
	if c.CreatedAt != nil {
		return RecurlyError{statusCode:400,Description:"Coupon Already created"}
	}
	return c.r.doCreate(&c,c.endpoint)
}

//Redeem a coupon on an account
func (c *Coupon) Redeem(account_code string, currency string) (error) {
	redemption := Redemption{AccountCode: account_code, Currency:currency}
	redemption.r = c.r
	return redemption.r.doCreate(&redemption,c.endpoint + "/" + c.CouponCode + "/redeem" )
}

//delete and adjustment
func (c *Coupon) Deactivate() (error) {
	return c.r.doDelete(c.endpoint + "/" + c.CouponCode)
}

type CouponList struct {
	Paging
	r *Recurly
	XMLName xml.Name `xml:"coupons"`
	Coupons []Coupon `xml:"coupon"`
}

//Get next set of Coupons
func (c *CouponList) Next() (bool) {
	if c.next != "" {
		v := url.Values{}
		v.Set("cursor",c.next)
		v.Set("per_page",c.perPage)
		*c,_ = c.r.GetCoupons(v)
	} else {
		return false
	}
	return true
}

//Get previous set of accounts
func (c *CouponList) Prev() ( bool) {
	if c.prev != "" {
		v := url.Values{}
		v.Set("cursor",c.prev)
		v.Set("per_page",c.perPage)
		*c,_ = c.r.GetCoupons(v)
	} else {
		return false
	}
	return true
}

//Go to start set of accounts
func (c *CouponList) Start() ( bool) {
	if c.prev != "" {
		v := url.Values{}
		v.Set("per_page",c.perPage)
		*c,_ = c.r.GetCoupons(v)
	} else {
		return false
	}
	return true
}

type LineItems struct {
	XMLName xml.Name `xml:"line_items"`
	Adjustment []Adjustment
}

type Invoice struct {
	XMLName xml.Name `xml:"invoice"`
	endpoint string
	r *Recurly
	Account *AccountStub `xml:"account,omitempty"`
	UUID string `xml:"uuid,omitempty"`
	State string `xml:"state,omitempty"`
	InvoiceNumber string `xml:"invoice_number,omitempty"`
	PONumber string `xml:"po_number,omitempty"`
	VATNumber string `xml:"vat_number,omitempty"`
	SubtotalInCents int `xml:"subtotal_in_cents,omitempty"`
	TaxInCents int `xml:"tax_in_cents,omitempty"`
	TotalInCents int `xml:"total_in_cents,omitempty"`
	Currency string `xml:"currency,omitempty"`
	CreatedAt *time.Time `xml:"created_at,omitempty"`
	LineItems []LineItems `xml:"line_items,omitempty"`
	//Transactions Transaction `xml:"transactions,omitempty"`
}

type InvoiceList struct {
	Paging
	r *Recurly
	XMLName xml.Name `xml:"invoices"`
	Invoices []Invoice `xml:"invoice"`
}


//Get next set of Coupons
func (i *InvoiceList) Next() (bool) {
	if i.next != "" {
		v := url.Values{}
		v.Set("cursor",i.next)
		v.Set("per_page",i.perPage)
		*i,_ = i.r.GetInvoices(v)
	} else {
		return false
	}
	return true
}

//Get previous set of accounts
func (i *InvoiceList) Prev() ( bool) {
	if i.prev != "" {
		v := url.Values{}
		v.Set("cursor",i.prev)
		v.Set("per_page",i.perPage)
		*i,_ = i.r.GetInvoices(v)
	} else {
		return false
	}
	return true
}

//Go to start set of accounts
func (i *InvoiceList) Start() ( bool) {
	if i.prev != "" {
		v := url.Values{}
		v.Set("per_page",i.perPage)
		*i,_ = i.r.GetInvoices(v)
	} else {
		return false
	}
	return true
}

type AccountInvoiceList struct {
	Paging
	r *Recurly
	XMLName xml.Name `xml:"invoices"`
	AccountCode string `xml:"-"`
	Invoices []Invoice `xml:"invoice"`
}


//Get next set of Coupons
func (a *AccountInvoiceList) Next() (bool) {
	if a.next != "" {
		v := url.Values{}
		v.Set("cursor",a.next)
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAccountInvoices(a.AccountCode,v)
	} else {
		return false
	}
	return true
}

//Get previous set of accounts
func (a *AccountInvoiceList) Prev() ( bool) {
	if a.prev != "" {
		v := url.Values{}
		v.Set("cursor",a.prev)
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAccountInvoices(a.AccountCode,v)
	} else {
		return false
	}
	return true
}

//Go to start set of accounts
func (a *AccountInvoiceList) Start() ( bool) {
	if a.prev != "" {
		v := url.Values{}
		v.Set("per_page",a.perPage)
		*a,_ = a.r.GetAccountInvoices(a.AccountCode,v)
	} else {
		return false
	}
	return true
}

type CurrencyArray struct {
	CurrencyList []Currency
}

func (c *CurrencyArray) AddCurrency(currency string, amount int) {
	newc := Currency{Amount:amount}
	newc.XMLName.Local = currency
	c.CurrencyList = append(c.CurrencyList, newc)
}

type Currency struct {
	XMLName xml.Name
	Amount int `xml:",innerxml"`
}


type Plan struct {
	XMLName xml.Name `xml:"plan"`
	endpoint string
	r *Recurly
	//AddOns *AccountStub `xml:"add_ons,omitempty"`
	Name string `xml:"name,omitempty"`
	PlanCode string `xml:"state,omitempty"`
	Description string `xml:"po_number,omitempty"`
	SuccessUrl string `xml:"invoice_number,omitempty"`
	CancelUrl string `xml:"vat_number,omitempty"`
	DisplayDontaionAmounts bool `xml:"subtotal_in_cents,omitempty"`
	DisplayQuantity bool `xml:"tax_in_cents,omitempty"`
	DisplayPhoneNumber bool `xml:"tax_in_cents,omitempty"`
	BypassHostedConfirmation bool `xml:"tax_in_cents,omitempty"`
	UnitName string `xml:"tax_in_cents,omitempty"`
	PaymentPageTOSLink string `xml:"total_in_cents,omitempty"`
	PlanIntervalLength int `xml:"currency,omitempty"`
	PlanIntervalUnit string `xml:"currency,omitempty"`
	AccountingCode string `xml:"currency,omitempty"`
	CreatedAt *time.Time `xml:"created_at,omitempty"`
	UnitAmountInCents Currency `xml:",omitempty"`
	SetupFeeInCents Currency `xml:",omitempty"`
}
/* end resource objects */


