// Copyright 2022 Wenfeng Zhou (zwf1094646850@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1beta1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	crt := "/tls/tls.crt"
	key := "/tls/tls.key"

	fmt.Printf("to find crt at `%s`\n", crt)
	fmt.Printf("to find key at `%s`\n", key)
	fmt.Printf("http server listen at :9999\n")

	http.HandleFunc("/", HandleMutate)
	err := http.ListenAndServeTLS(":443", crt, key, nil)
	if err != nil {
		panic(err)
	}
}

type JSONPatchEntry struct {
	OP    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func HandleMutate(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		fmt.Printf("[E] error -1, request content type is not json\n")
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	admissionReview := &admissionv1beta1.AdmissionReview{}

	if err := json.NewDecoder(r.Body).Decode(admissionReview); err != nil {
		fmt.Printf("[E] error 1, %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to decode admission review, %v", err)
		return
	}

	// unmarshal the pod from the AdmissionRequest
	pod := &corev1.Pod{}
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, pod); err != nil {
		fmt.Printf("[E] error 2, %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to unmarshal pod, %v", err)
		return
	}

	// build json patch
	// Notice: 这里 patch label 的时候，如果内容中期望包含 '/', 需要手动将其编码为 `~1`, 参见: https://stackoverflow.com/questions/65887327/patch-kubernetes-label-with-character
	patch := []JSONPatchEntry{
		JSONPatchEntry{
			OP:    "replace",
			Path:  "/metadata/labels/app.demeter.io~1name",
			Value: "ok",
		},
	}

	patchBytes, err := json.Marshal(&patch)
	if err != nil {
		fmt.Printf("[E] error 3, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to json marshal patch, %v", err)
		return
	}

	patchType := admissionv1beta1.PatchTypeJSONPatch

	// build admission response
	admissionResponse := &admissionv1beta1.AdmissionResponse{
		UID:       admissionReview.Request.UID,
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}

	respAdmissionReview := &admissionv1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       admissionReview.Kind,
			APIVersion: admissionReview.APIVersion,
		},
		Response: admissionResponse,
	}

	b, err := json.Marshal(respAdmissionReview)
	if err != nil {
		fmt.Printf("[E] error 5, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to marshal admission review, %v", err)
		return
	}
	fmt.Printf("[I] webhook ok, %s\n", string(b))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
