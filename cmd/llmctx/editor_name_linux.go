package main

func editor() editorPreset {
	name := editorName()
	res := getCommonEditorPreset(name)
	if res != nil {
		return *res
	}

	return editorPreset{
		Binary: "edit",
	}
}
