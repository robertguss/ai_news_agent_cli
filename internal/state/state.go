package state

import (
        "encoding/json"
        "errors"
        "os"
        "path/filepath"
        "time"
)

type ArticleRef struct {
        ID           int64  `json:"id"`
        URL          string `json:"url"`
        Title        string `json:"title"`
        StoryGroupID string `json:"story_group_id"`
}

type ViewState struct {
        Timestamp time.Time             `json:"timestamp"`
        Articles  map[string]ArticleRef `json:"articles"`
}

var pathFunc = func() (string, error) {
        home, err := os.UserHomeDir()
        if err != nil {
                return "", err
        }
        return filepath.Join(home, ".ai-news-state.json"), nil
}

func GetPathFunc() func() (string, error) {
        return pathFunc
}

func SetPathFunc(f func() (string, error)) {
        pathFunc = f
}

func Path() (string, error) {
        return pathFunc()
}

func Load() (*ViewState, error) {
        p, err := Path()
        if err != nil {
                return nil, err
        }
        
        f, err := os.Open(p)
        if errors.Is(err, os.ErrNotExist) {
                return &ViewState{Articles: map[string]ArticleRef{}}, nil
        }
        if err != nil {
                return nil, err
        }
        defer f.Close()
        
        var vs ViewState
        if err := json.NewDecoder(f).Decode(&vs); err != nil {
                return nil, err
        }
        
        return &vs, nil
}

func Save(vs *ViewState) error {
        p, err := Path()
        if err != nil {
                return err
        }
        
        tmp := p + ".tmp"
        f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
        if err != nil {
                return err
        }
        
        enc := json.NewEncoder(f)
        enc.SetIndent("", "  ")
        if err := enc.Encode(vs); err != nil {
                f.Close()
                return err
        }
        
        if err := f.Close(); err != nil {
                return err
        }
        
        return os.Rename(tmp, p)
}
