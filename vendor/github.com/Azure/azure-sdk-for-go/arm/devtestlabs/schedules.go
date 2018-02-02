package devtestlabs

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/validation"
	"net/http"
)

// SchedulesClient is the the DevTest Labs Client.
type SchedulesClient struct {
	ManagementClient
}

// NewSchedulesClient creates an instance of the SchedulesClient client.
func NewSchedulesClient(subscriptionID string) SchedulesClient {
	return NewSchedulesClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewSchedulesClientWithBaseURI creates an instance of the SchedulesClient client.
func NewSchedulesClientWithBaseURI(baseURI string, subscriptionID string) SchedulesClient {
	return SchedulesClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// CreateOrUpdate create or replace an existing schedule.
//
// resourceGroupName is the name of the resource group. labName is the name of the lab. name is the name of the
// schedule. schedule is a schedule.
func (client SchedulesClient) CreateOrUpdate(resourceGroupName string, labName string, name string, schedule Schedule) (result Schedule, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: schedule,
			Constraints: []validation.Constraint{{Target: "schedule.ScheduleProperties", Name: validation.Null, Rule: true, Chain: nil}}}}); err != nil {
		return result, validation.NewErrorWithValidationError(err, "devtestlabs.SchedulesClient", "CreateOrUpdate")
	}

	req, err := client.CreateOrUpdatePreparer(resourceGroupName, labName, name, schedule)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "CreateOrUpdate", nil, "Failure preparing request")
		return
	}

	resp, err := client.CreateOrUpdateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "CreateOrUpdate", resp, "Failure sending request")
		return
	}

	result, err = client.CreateOrUpdateResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "CreateOrUpdate", resp, "Failure responding to request")
	}

	return
}

// CreateOrUpdatePreparer prepares the CreateOrUpdate request.
func (client SchedulesClient) CreateOrUpdatePreparer(resourceGroupName string, labName string, name string, schedule Schedule) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"labName":           autorest.Encode("path", labName),
		"name":              autorest.Encode("path", name),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-05-15"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsJSON(),
		autorest.AsPut(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DevTestLab/labs/{labName}/schedules/{name}", pathParameters),
		autorest.WithJSON(schedule),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// CreateOrUpdateSender sends the CreateOrUpdate request. The method will close the
// http.Response Body if it receives an error.
func (client SchedulesClient) CreateOrUpdateSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// CreateOrUpdateResponder handles the response to the CreateOrUpdate request. The method always
// closes the http.Response Body.
func (client SchedulesClient) CreateOrUpdateResponder(resp *http.Response) (result Schedule, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// Delete delete schedule.
//
// resourceGroupName is the name of the resource group. labName is the name of the lab. name is the name of the
// schedule.
func (client SchedulesClient) Delete(resourceGroupName string, labName string, name string) (result autorest.Response, err error) {
	req, err := client.DeletePreparer(resourceGroupName, labName, name)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Delete", nil, "Failure preparing request")
		return
	}

	resp, err := client.DeleteSender(req)
	if err != nil {
		result.Response = resp
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Delete", resp, "Failure sending request")
		return
	}

	result, err = client.DeleteResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Delete", resp, "Failure responding to request")
	}

	return
}

// DeletePreparer prepares the Delete request.
func (client SchedulesClient) DeletePreparer(resourceGroupName string, labName string, name string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"labName":           autorest.Encode("path", labName),
		"name":              autorest.Encode("path", name),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-05-15"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsDelete(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DevTestLab/labs/{labName}/schedules/{name}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// DeleteSender sends the Delete request. The method will close the
// http.Response Body if it receives an error.
func (client SchedulesClient) DeleteSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// DeleteResponder handles the response to the Delete request. The method always
// closes the http.Response Body.
func (client SchedulesClient) DeleteResponder(resp *http.Response) (result autorest.Response, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusNoContent),
		autorest.ByClosing())
	result.Response = resp
	return
}

// Execute execute a schedule. This operation can take a while to complete. This method may poll for completion.
// Polling can be canceled by passing the cancel channel argument. The channel will be used to cancel polling and any
// outstanding HTTP requests.
//
// resourceGroupName is the name of the resource group. labName is the name of the lab. name is the name of the
// schedule.
func (client SchedulesClient) Execute(resourceGroupName string, labName string, name string, cancel <-chan struct{}) (<-chan autorest.Response, <-chan error) {
	resultChan := make(chan autorest.Response, 1)
	errChan := make(chan error, 1)
	go func() {
		var err error
		var result autorest.Response
		defer func() {
			if err != nil {
				errChan <- err
			}
			resultChan <- result
			close(resultChan)
			close(errChan)
		}()
		req, err := client.ExecutePreparer(resourceGroupName, labName, name, cancel)
		if err != nil {
			err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Execute", nil, "Failure preparing request")
			return
		}

		resp, err := client.ExecuteSender(req)
		if err != nil {
			result.Response = resp
			err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Execute", resp, "Failure sending request")
			return
		}

		result, err = client.ExecuteResponder(resp)
		if err != nil {
			err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Execute", resp, "Failure responding to request")
		}
	}()
	return resultChan, errChan
}

// ExecutePreparer prepares the Execute request.
func (client SchedulesClient) ExecutePreparer(resourceGroupName string, labName string, name string, cancel <-chan struct{}) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"labName":           autorest.Encode("path", labName),
		"name":              autorest.Encode("path", name),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-05-15"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DevTestLab/labs/{labName}/schedules/{name}/execute", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{Cancel: cancel})
}

// ExecuteSender sends the Execute request. The method will close the
// http.Response Body if it receives an error.
func (client SchedulesClient) ExecuteSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client,
		req,
		azure.DoPollForAsynchronous(client.PollingDelay))
}

// ExecuteResponder handles the response to the Execute request. The method always
// closes the http.Response Body.
func (client SchedulesClient) ExecuteResponder(resp *http.Response) (result autorest.Response, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusAccepted),
		autorest.ByClosing())
	result.Response = resp
	return
}

// Get get schedule.
//
// resourceGroupName is the name of the resource group. labName is the name of the lab. name is the name of the
// schedule. expand is specify the $expand query. Example: 'properties($select=status)'
func (client SchedulesClient) Get(resourceGroupName string, labName string, name string, expand string) (result Schedule, err error) {
	req, err := client.GetPreparer(resourceGroupName, labName, name, expand)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Get", nil, "Failure preparing request")
		return
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Get", resp, "Failure sending request")
		return
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Get", resp, "Failure responding to request")
	}

	return
}

// GetPreparer prepares the Get request.
func (client SchedulesClient) GetPreparer(resourceGroupName string, labName string, name string, expand string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"labName":           autorest.Encode("path", labName),
		"name":              autorest.Encode("path", name),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-05-15"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}
	if len(expand) > 0 {
		queryParameters["$expand"] = autorest.Encode("query", expand)
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DevTestLab/labs/{labName}/schedules/{name}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client SchedulesClient) GetSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client SchedulesClient) GetResponder(resp *http.Response) (result Schedule, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// List list schedules in a given lab.
//
// resourceGroupName is the name of the resource group. labName is the name of the lab. expand is specify the $expand
// query. Example: 'properties($select=status)' filter is the filter to apply to the operation. top is the maximum
// number of resources to return from the operation. orderby is the ordering expression for the results, using OData
// notation.
func (client SchedulesClient) List(resourceGroupName string, labName string, expand string, filter string, top *int32, orderby string) (result ResponseWithContinuationSchedule, err error) {
	req, err := client.ListPreparer(resourceGroupName, labName, expand, filter, top, orderby)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "List", nil, "Failure preparing request")
		return
	}

	resp, err := client.ListSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "List", resp, "Failure sending request")
		return
	}

	result, err = client.ListResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "List", resp, "Failure responding to request")
	}

	return
}

// ListPreparer prepares the List request.
func (client SchedulesClient) ListPreparer(resourceGroupName string, labName string, expand string, filter string, top *int32, orderby string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"labName":           autorest.Encode("path", labName),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-05-15"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}
	if len(expand) > 0 {
		queryParameters["$expand"] = autorest.Encode("query", expand)
	}
	if len(filter) > 0 {
		queryParameters["$filter"] = autorest.Encode("query", filter)
	}
	if top != nil {
		queryParameters["$top"] = autorest.Encode("query", *top)
	}
	if len(orderby) > 0 {
		queryParameters["$orderby"] = autorest.Encode("query", orderby)
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DevTestLab/labs/{labName}/schedules", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// ListSender sends the List request. The method will close the
// http.Response Body if it receives an error.
func (client SchedulesClient) ListSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// ListResponder handles the response to the List request. The method always
// closes the http.Response Body.
func (client SchedulesClient) ListResponder(resp *http.Response) (result ResponseWithContinuationSchedule, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListNextResults retrieves the next set of results, if any.
func (client SchedulesClient) ListNextResults(lastResults ResponseWithContinuationSchedule) (result ResponseWithContinuationSchedule, err error) {
	req, err := lastResults.ResponseWithContinuationSchedulePreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "List", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}

	resp, err := client.ListSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "List", resp, "Failure sending next results request")
	}

	result, err = client.ListResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "List", resp, "Failure responding to next results request")
	}

	return
}

// ListComplete gets all elements from the list without paging.
func (client SchedulesClient) ListComplete(resourceGroupName string, labName string, expand string, filter string, top *int32, orderby string, cancel <-chan struct{}) (<-chan Schedule, <-chan error) {
	resultChan := make(chan Schedule)
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			close(resultChan)
			close(errChan)
		}()
		list, err := client.List(resourceGroupName, labName, expand, filter, top, orderby)
		if err != nil {
			errChan <- err
			return
		}
		if list.Value != nil {
			for _, item := range *list.Value {
				select {
				case <-cancel:
					return
				case resultChan <- item:
					// Intentionally left blank
				}
			}
		}
		for list.NextLink != nil {
			list, err = client.ListNextResults(list)
			if err != nil {
				errChan <- err
				return
			}
			if list.Value != nil {
				for _, item := range *list.Value {
					select {
					case <-cancel:
						return
					case resultChan <- item:
						// Intentionally left blank
					}
				}
			}
		}
	}()
	return resultChan, errChan
}

// ListApplicable lists all applicable schedules
//
// resourceGroupName is the name of the resource group. labName is the name of the lab. name is the name of the
// schedule.
func (client SchedulesClient) ListApplicable(resourceGroupName string, labName string, name string) (result ResponseWithContinuationSchedule, err error) {
	req, err := client.ListApplicablePreparer(resourceGroupName, labName, name)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "ListApplicable", nil, "Failure preparing request")
		return
	}

	resp, err := client.ListApplicableSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "ListApplicable", resp, "Failure sending request")
		return
	}

	result, err = client.ListApplicableResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "ListApplicable", resp, "Failure responding to request")
	}

	return
}

// ListApplicablePreparer prepares the ListApplicable request.
func (client SchedulesClient) ListApplicablePreparer(resourceGroupName string, labName string, name string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"labName":           autorest.Encode("path", labName),
		"name":              autorest.Encode("path", name),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-05-15"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DevTestLab/labs/{labName}/schedules/{name}/listApplicable", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// ListApplicableSender sends the ListApplicable request. The method will close the
// http.Response Body if it receives an error.
func (client SchedulesClient) ListApplicableSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// ListApplicableResponder handles the response to the ListApplicable request. The method always
// closes the http.Response Body.
func (client SchedulesClient) ListApplicableResponder(resp *http.Response) (result ResponseWithContinuationSchedule, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListApplicableNextResults retrieves the next set of results, if any.
func (client SchedulesClient) ListApplicableNextResults(lastResults ResponseWithContinuationSchedule) (result ResponseWithContinuationSchedule, err error) {
	req, err := lastResults.ResponseWithContinuationSchedulePreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "ListApplicable", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}

	resp, err := client.ListApplicableSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "ListApplicable", resp, "Failure sending next results request")
	}

	result, err = client.ListApplicableResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "ListApplicable", resp, "Failure responding to next results request")
	}

	return
}

// ListApplicableComplete gets all elements from the list without paging.
func (client SchedulesClient) ListApplicableComplete(resourceGroupName string, labName string, name string, cancel <-chan struct{}) (<-chan Schedule, <-chan error) {
	resultChan := make(chan Schedule)
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			close(resultChan)
			close(errChan)
		}()
		list, err := client.ListApplicable(resourceGroupName, labName, name)
		if err != nil {
			errChan <- err
			return
		}
		if list.Value != nil {
			for _, item := range *list.Value {
				select {
				case <-cancel:
					return
				case resultChan <- item:
					// Intentionally left blank
				}
			}
		}
		for list.NextLink != nil {
			list, err = client.ListApplicableNextResults(list)
			if err != nil {
				errChan <- err
				return
			}
			if list.Value != nil {
				for _, item := range *list.Value {
					select {
					case <-cancel:
						return
					case resultChan <- item:
						// Intentionally left blank
					}
				}
			}
		}
	}()
	return resultChan, errChan
}

// Update modify properties of schedules.
//
// resourceGroupName is the name of the resource group. labName is the name of the lab. name is the name of the
// schedule. schedule is a schedule.
func (client SchedulesClient) Update(resourceGroupName string, labName string, name string, schedule ScheduleFragment) (result Schedule, err error) {
	req, err := client.UpdatePreparer(resourceGroupName, labName, name, schedule)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Update", nil, "Failure preparing request")
		return
	}

	resp, err := client.UpdateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Update", resp, "Failure sending request")
		return
	}

	result, err = client.UpdateResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "devtestlabs.SchedulesClient", "Update", resp, "Failure responding to request")
	}

	return
}

// UpdatePreparer prepares the Update request.
func (client SchedulesClient) UpdatePreparer(resourceGroupName string, labName string, name string, schedule ScheduleFragment) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"labName":           autorest.Encode("path", labName),
		"name":              autorest.Encode("path", name),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-05-15"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsJSON(),
		autorest.AsPatch(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DevTestLab/labs/{labName}/schedules/{name}", pathParameters),
		autorest.WithJSON(schedule),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// UpdateSender sends the Update request. The method will close the
// http.Response Body if it receives an error.
func (client SchedulesClient) UpdateSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// UpdateResponder handles the response to the Update request. The method always
// closes the http.Response Body.
func (client SchedulesClient) UpdateResponder(resp *http.Response) (result Schedule, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
