package exchange

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/open-horizon/anax/businesspolicy"
	"github.com/open-horizon/anax/externalpolicy"
	"strings"
	"time"
)

// The node and service policy objects in the exchange are identical to the external policy object
// supported by the node/policy API, so it is embedded in the ExchangePolicy object.
type ExchangePolicy struct {
	externalpolicy.ExternalPolicy
	LastUpdated string `json:"lastUpdated,omitempty"`
}

func (e ExchangePolicy) String() string {
	return fmt.Sprintf("%v, "+
		"LastUpdated: %v",
		e.ExternalPolicy, e.LastUpdated)
}

func (e ExchangePolicy) ShortString() string {
	return e.String()
}

func (e *ExchangePolicy) GetExternalPolicy() externalpolicy.ExternalPolicy {
	return e.ExternalPolicy
}

func (e *ExchangePolicy) GetLastUpdated() string {
	return e.LastUpdated
}

// the exchange business policy
type ExchangeBusinessPolicy struct {
	businesspolicy.BusinessPolicy
	Created     string `json:"created,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}

func (e ExchangeBusinessPolicy) String() string {
	return fmt.Sprintf("%v, "+
		"Created: %v, "+
		"LastUpdated: %v",
		e.BusinessPolicy, e.Created, e.LastUpdated)
}

func (e ExchangeBusinessPolicy) ShortString() string {
	return e.String()
}

func (e *ExchangeBusinessPolicy) GetBusinessPolicy() businesspolicy.BusinessPolicy {
	return e.BusinessPolicy
}

func (e *ExchangeBusinessPolicy) GetLastUpdated() string {
	return e.LastUpdated
}

func (e *ExchangeBusinessPolicy) GetCreated() string {
	return e.Created
}

type GetBusinessPolicyResponse struct {
	BusinessPolicy map[string]ExchangeBusinessPolicy `json:"businessPolicy,omitempty"` // map of all defined business policies
	LastIndex      int                               `json:"lastIndex.omitempty"`
}

// Structs and types for working with business policy based exchange searches
type SearchExchBusinessPolRequest struct {
	NodeOrgIds   []string `json:"nodeOrgids,omitempty"`
	ChangedSince uint64   `json:"changedSince"`
}

func (a SearchExchBusinessPolRequest) String() string {
	return fmt.Sprintf("NodeOrgIds: %v, ChangedSince: %v", a.NodeOrgIds, a.ChangedSince)
}

type SearchExchBusinessPolResponse struct {
	Devices   []SearchResultDevice `json:"nodes"`
	LastIndex int                  `json:"lastIndex"`
}

func (r SearchExchBusinessPolResponse) String() string {
	return fmt.Sprintf("Devices: %v, LastIndex: %v", r.Devices, r.LastIndex)
}

// Retrieve the node policy object from the exchange. The input device Id is assumed to be prefixed with its org.
func GetNodePolicy(ec ExchangeContext, deviceId string) (*ExchangePolicy, error) {
	glog.V(3).Infof(rpclogString(fmt.Sprintf("getting node policy for %v.", deviceId)))

	// Get the node policy object. There should only be 1.
	var resp interface{}
	resp = new(ExchangePolicy)

	targetURL := fmt.Sprintf("%vorgs/%v/nodes/%v/policy", ec.GetExchangeURL(), GetOrg(deviceId), GetId(deviceId))

	retryCount := ec.GetHTTPFactory().RetryCount
	retryInterval := ec.GetHTTPFactory().GetRetryInterval()
	for {
		if err, tpErr := InvokeExchange(ec.GetHTTPFactory().NewHTTPClient(nil), "GET", targetURL, ec.GetExchangeId(), ec.GetExchangeToken(), nil, &resp); err != nil {
			glog.Errorf(rpclogString(fmt.Sprintf(err.Error())))
			return nil, err
		} else if tpErr != nil {
			glog.Warningf(rpclogString(fmt.Sprintf(tpErr.Error())))
			if ec.GetHTTPFactory().RetryCount == 0 {
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			} else if retryCount == 0 {
				return nil, fmt.Errorf("Exceeded %v retries for error: %v", ec.GetHTTPFactory().RetryCount, tpErr)
			} else {
				retryCount--
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}
		} else {
			glog.V(5).Infof(rpclogString(fmt.Sprintf("returning node policy %v for %v.", resp, deviceId)))
			nodePolicy := resp.(*ExchangePolicy)
			if nodePolicy.GetLastUpdated() == "" {
				return nil, nil
			} else {
				return nodePolicy, nil
			}
		}
	}

}

// Write an updated node policy to the exchange.
func PutNodePolicy(ec ExchangeContext, deviceId string, ep *ExchangePolicy) (*PutDeviceResponse, error) {
	// create PUT body
	var resp interface{}
	resp = new(PutDeviceResponse)
	targetURL := fmt.Sprintf("%vorgs/%v/nodes/%v/policy", ec.GetExchangeURL(), GetOrg(deviceId), GetId(deviceId))

	retryCount := ec.GetHTTPFactory().RetryCount
	retryInterval := ec.GetHTTPFactory().GetRetryInterval()
	for {
		if err, tpErr := InvokeExchange(ec.GetHTTPFactory().NewHTTPClient(nil), "PUT", targetURL, ec.GetExchangeId(), ec.GetExchangeToken(), ep, &resp); err != nil {
			return nil, err
		} else if tpErr != nil {
			glog.Warningf(rpclogString(fmt.Sprintf(tpErr.Error())))
			if ec.GetHTTPFactory().RetryCount == 0 {
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			} else if retryCount == 0 {
				return nil, fmt.Errorf("Exceeded %v retries for error: %v", ec.GetHTTPFactory().RetryCount, tpErr)
			} else {
				retryCount--
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}
		} else {
			glog.V(3).Infof(rpclogString(fmt.Sprintf("put device policy for %v to exchange %v", deviceId, ep)))
			return resp.(*PutDeviceResponse), nil
		}
	}
}

// Delete node policy from the exchange.
// Return nil if the policy is deleted or does not exist.
func DeleteNodePolicy(ec ExchangeContext, deviceId string) error {
	// create PUT body
	var resp interface{}
	resp = new(PostDeviceResponse)
	targetURL := fmt.Sprintf("%vorgs/%v/nodes/%v/policy", ec.GetExchangeURL(), GetOrg(deviceId), GetId(deviceId))

	retryCount := ec.GetHTTPFactory().RetryCount
	retryInterval := ec.GetHTTPFactory().GetRetryInterval()
	for {
		if err, tpErr := InvokeExchange(ec.GetHTTPFactory().NewHTTPClient(nil), "DELETE", targetURL, ec.GetExchangeId(), ec.GetExchangeToken(), nil, &resp); err != nil && !strings.Contains(err.Error(), "status: 404") {
			return err
		} else if tpErr != nil {
			glog.Warningf(rpclogString(fmt.Sprintf(tpErr.Error())))
			if ec.GetHTTPFactory().RetryCount == 0 {
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			} else if retryCount == 0 {
				return fmt.Errorf("Exceeded %v retries for error: %v", ec.GetHTTPFactory().RetryCount, tpErr)
			} else {
				retryCount--
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}
		} else {
			glog.V(3).Infof(rpclogString(fmt.Sprintf("deleted device policy for %v from the exchange.", deviceId)))
			return nil
		}
	}
}

// Get all the business policy metadata for a specific organization, and policy if specified.
func GetBusinessPolicies(ec ExchangeContext, org string, policy_id string) (map[string]ExchangeBusinessPolicy, error) {

	if policy_id == "" {
		glog.V(3).Infof(rpclogString(fmt.Sprintf("getting business policy for %v", org)))
	} else {
		glog.V(3).Infof(rpclogString(fmt.Sprintf("getting business policy for %v/%v", org, policy_id)))
	}

	var resp interface{}
	resp = new(GetBusinessPolicyResponse)

	// Search the exchange for the business policy definitions
	targetURL := ""
	if policy_id == "" {
		targetURL = fmt.Sprintf("%vorgs/%v/business/policies", ec.GetExchangeURL(), org)
	} else {
		targetURL = fmt.Sprintf("%vorgs/%v/business/policies/%v", ec.GetExchangeURL(), org, policy_id)
	}

	retryCount := ec.GetHTTPFactory().RetryCount
	retryInterval := ec.GetHTTPFactory().GetRetryInterval()
	for {
		if err, tpErr := InvokeExchange(ec.GetHTTPFactory().NewHTTPClient(nil), "GET", targetURL, ec.GetExchangeId(), ec.GetExchangeToken(), nil, &resp); err != nil {
			glog.Errorf(rpclogString(fmt.Sprintf(err.Error())))
			return nil, err
		} else if tpErr != nil {
			glog.Warningf(rpclogString(fmt.Sprintf(tpErr.Error())))
			if ec.GetHTTPFactory().RetryCount == 0 {
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			} else if retryCount == 0 {
				return nil, fmt.Errorf("Exceeded %v retries for error: %v", ec.GetHTTPFactory().RetryCount, tpErr)
			} else {
				retryCount--
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}
		} else {
			pols := resp.(*GetBusinessPolicyResponse).BusinessPolicy
			glog.V(3).Infof(rpclogString(fmt.Sprintf("found business policy for %v, %v", org, pols)))
			return pols, nil
		}
	}
}
